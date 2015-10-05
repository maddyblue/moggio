package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mjibson/appstats"
	"github.com/mjibson/goon"
	"github.com/mjibson/mog/models"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/user"
)

var (
	templates = template.Must(template.ParseGlob("templates/*.html"))
	errCSRF   = fmt.Errorf("bad csrf")
)

func init() {
	http.HandleFunc("/token", F(TokenForm))
	http.HandleFunc("/token/get", F(TokenGet))
	http.HandleFunc("/api/username", F(Username))
	http.HandleFunc("/api/source/set", F(SetSource))
	http.HandleFunc("/api/source/get", F(GetSource))
	http.HandleFunc("/api/source/delete", F(DeleteSource))
}

func F(f func(c context.Context, w http.ResponseWriter, r *http.Request, u *User, g *goon.Goon) (interface{}, error)) http.HandlerFunc {
	return appstats.NewHandlerFunc(func(c context.Context, w http.ResponseWriter, r *http.Request) {
		var u *User
		au := user.Current(c)
		if (au == nil || au.ID == "") && !strings.HasPrefix(r.URL.Path, "/api/") {
			url, err := user.LoginURL(c, r.URL.String())
			if err != nil {
				serveError(w, err)
				return
			}
			http.Redirect(w, r, url, 302)
			return
		}
		g := goon.FromContext(c)
		if au != nil && au.ID != "" {
			u = &User{ID: au.ID}
			if err := g.Get(u); err != nil && err != datastore.ErrNoSuchEntity {
				serveError(w, err)
				return
			}
			if now := time.Now(); u.CSRFExpire.Before(now) {
				b := make([]byte, 32)
				if _, err := rand.Read(b); err != nil {
					serveError(w, err)
					return
				}
				u.Email = au.Email
				u.CSRF = base64.URLEncoding.EncodeToString(b)
				u.CSRFExpire = now.Add(time.Hour)
				if _, err := g.Put(u); err != nil {
					serveError(w, err)
					return
				}
			}
		} else {
			t := &Token{
				ID: r.FormValue("token"),
			}
			if t.ID == "" {
				serveError(w, fmt.Errorf("no token"))
				return
			}
			if err := g.Get(t); err != nil {
				serveError(w, err)
				return
			}
			u = &User{
				ID: t.User.StringID(),
			}
			if err := g.Get(u); err != nil {
				serveError(w, err)
				return
			}
		}
		if err := r.ParseForm(); err != nil {
			serveError(w, err)
			return
		}
		d, err := f(c, w, r, u, g)
		if err != nil {
			serveError(w, err)
			return
		}
		if d == nil {
			return
		}
		b, err := json.Marshal(d)
		if err != nil {
			serveError(w, err)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(b)
	})
}

func serveError(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

type User struct {
	ID         string `datastore:"-" goon:"id"`
	Email      string
	CSRF       string
	CSRFExpire time.Time
}

type Token struct {
	ID       string `datastore:"-" goon:"id"`
	User     *datastore.Key
	Name     string
	Issued   time.Time
	LastUsed time.Time
}

type Protocol struct {
	ID   string         `goon:"id"`
	User *datastore.Key `goon:"parent"`
}

type Source struct {
	ID       string         `datastore:"-" goon:"id"`
	Protocol *datastore.Key `datastore:"-" goon:"parent"`
	Blob     []byte
}

func TokenForm(c context.Context, w http.ResponseWriter, r *http.Request, u *User, g *goon.Goon) (interface{}, error) {
	data := struct {
		User     *User
		Redirect string
		Hostname string
	}{
		u,
		r.FormValue("redirect"),
		r.FormValue("hostname"),
	}
	err := templates.ExecuteTemplate(w, "token.html", data)
	return nil, err
}

func validCSRF(csrf string, user *User) bool {
	if user.CSRF == "" || user.CSRFExpire.Before(time.Now()) {
		return false
	}
	return csrf == user.CSRF
}

func TokenGet(c context.Context, w http.ResponseWriter, r *http.Request, u *User, g *goon.Goon) (interface{}, error) {
	if !validCSRF(r.FormValue("csrf"), u) {
		return nil, errCSRF
	}
	redir, err := url.Parse(r.FormValue("redirect"))
	if err != nil {
		return nil, err
	}
	uk, err := g.KeyError(u)
	if err != nil {
		return nil, err
	}
	b := make([]byte, 64)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	token := &Token{
		ID:     base64.URLEncoding.EncodeToString(b),
		User:   uk,
		Name:   r.FormValue("hostname"),
		Issued: time.Now(),
	}
	if _, err := g.Put(token); err != nil {
		return nil, err
	}
	values := redir.Query()
	values.Add("token", token.ID)
	redir.RawQuery = values.Encode()
	http.Redirect(w, r, redir.String(), 302)
	return nil, nil
}

func Username(c context.Context, w http.ResponseWriter, r *http.Request, u *User, g *goon.Goon) (interface{}, error) {
	return u.Email, nil
}

func SetSource(c context.Context, w http.ResponseWriter, r *http.Request, u *User, g *goon.Goon) (interface{}, error) {
	var ss []*models.Source
	if err := json.NewDecoder(r.Body).Decode(&ss); err != nil {
		return nil, err
	}
	for _, s := range ss {
		p := &Protocol{
			ID:   s.Protocol,
			User: g.Key(u),
		}
		src := Source{
			ID:       s.Name,
			Protocol: g.Key(p),
			Blob:     s.Blob,
		}
		log.Println(s.Name, s.Protocol, len(s.Blob))
		_, err := g.Put(&src)
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func GetSource(c context.Context, w http.ResponseWriter, r *http.Request, u *User, g *goon.Goon) (interface{}, error) {
	var ss []*Source
	k := g.Kind(&Source{})
	q := datastore.NewQuery(k).Ancestor(g.Key(u))
	if _, err := g.GetAll(q, &ss); err != nil {
		return nil, err
	}
	ret := make([]*models.Source, len(ss))
	for i, s := range ss {
		ret[i] = &models.Source{
			Protocol: s.Protocol.StringID(),
			Name:     s.ID,
			Blob:     s.Blob,
		}
	}
	return ret, nil
}

func DeleteSource(c context.Context, w http.ResponseWriter, r *http.Request, u *User, g *goon.Goon) (interface{}, error) {
	var d models.Delete
	if err := json.NewDecoder(r.Body).Decode(&d); err != nil {
		return nil, err
	}
	p := &Protocol{
		ID:   d.Protocol,
		User: g.Key(u),
	}
	s := &Source{
		ID:       d.Name,
		Protocol: g.Key(p),
	}
	err := g.Delete(g.Key(s))
	return nil, err
}
