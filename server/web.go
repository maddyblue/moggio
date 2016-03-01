package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mjibson/moggio/protocol"
	"golang.org/x/net/websocket"
)

var MoggioVersion string

var indexHTML []byte

func (srv *Server) GetMux(devMode bool) *http.ServeMux {
	var err error
	webFS := FS(devMode)
	if devMode {
		log.Println("using local web assets")
	}
	index, err := webFS.Open("/static/index.html")
	if err != nil {
		log.Fatal(err)
	}
	b, err := ioutil.ReadAll(index)
	if err != nil {
		log.Fatal(err)
	}
	tmpl := template.Must(template.New("").Parse(string(b)))
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, struct {
		Version string
	}{
		MoggioVersion,
	}); err != nil {
		log.Fatal(err)
	}
	indexHTML = buf.Bytes()
	router := httprouter.New()
	router.GET("/api/cmd/:cmd", JSON(srv.Cmd))
	router.GET("/api/data/:type", JSON(srv.Data))
	router.GET("/api/oauth/:protocol", srv.OAuth)
	router.POST("/api/cmd/:cmd", JSON(srv.Cmd))
	router.POST("/api/queue/change", JSON(srv.QueueChange))
	router.POST("/api/playlist/change/:playlist", JSON(srv.PlaylistChange))
	router.POST("/api/protocol/add", JSON(srv.ProtocolAdd))
	router.POST("/api/protocol/remove", JSON(srv.ProtocolRemove))
	router.POST("/api/protocol/refresh", JSON(srv.ProtocolRefresh))

	// Needs POST from local moggio. Needs GET from App Engine redirect.
	router.GET("/api/token/register", srv.TokenRegister)
	router.POST("/api/token/register", srv.TokenRegister)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(webFS))
	mux.HandleFunc("/", Index)
	mux.Handle("/api/", router)
	mux.Handle("/ws/", websocket.Handler(srv.WebSocket))
	return mux
}

// ListenAndServe listens on the TCP network address addr and then calls
// Serve to handle requests on incoming connections.
func (srv *Server) ListenAndServe(addr string, devMode bool) error {
	mux := srv.GetMux(devMode)
	log.Println("moggio: listening on", addr)
	return http.ListenAndServe(addr, mux)
}

func Index(w http.ResponseWriter, r *http.Request) {
	w.Write(indexHTML)
}

func serveError(w http.ResponseWriter, err error) {
	log.Println(err)
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func JSON(h func(io.Reader, url.Values, httprouter.Params) (interface{}, error)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if err := r.ParseForm(); err != nil {
			serveError(w, err)
			return
		}
		d, err := h(r.Body, r.Form, ps)
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
	}
}

func (srv *Server) OAuth(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	done := make(chan error)
	srv.ch <- cmdAddOAuth{
		name: ps.ByName("protocol"),
		r:    r,
		done: done,
	}
	err := <-done
	if err != nil {
		serveError(w, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (srv *Server) Data(body io.Reader, form url.Values, ps httprouter.Params) (interface{}, error) {
	ch := make(chan *waitData)
	srv.ch <- cmdWaitData{
		wt:   waitType(ps.ByName("type")),
		done: ch,
	}
	return <-ch, nil
}

func (srv *Server) Cmd(body io.Reader, form url.Values, ps httprouter.Params) (interface{}, error) {
	switch cmd := ps.ByName("cmd"); cmd {
	case "play":
		srv.ch <- cmdPlay
	case "stop":
		srv.ch <- cmdStop
	case "next":
		srv.ch <- cmdNext
	case "prev":
		srv.ch <- cmdPrev
	case "pause":
		srv.ch <- cmdPause
	case "play_idx":
		i, err := strconv.Atoi(form.Get("idx"))
		if err != nil {
			return nil, err
		}
		srv.ch <- cmdPlayIdx(i)
	case "play_track":
		var uid string
		if err := json.NewDecoder(body).Decode(&uid); err != nil {
			return nil, err
		}
		srv.ch <- cmdPlayTrack(uid)
	case "random":
		srv.ch <- cmdRandom
	case "repeat":
		srv.ch <- cmdRepeat
	case "seek":
		d, err := time.ParseDuration(form.Get("pos"))
		if err != nil {
			return nil, err
		}
		srv.ch <- cmdSeek(d)
	case "min_duration":
		d, err := time.ParseDuration(form.Get("d"))
		if err != nil {
			return nil, err
		}
		srv.ch <- cmdMinDuration(d)
	default:
		return nil, fmt.Errorf("unknown command: %v", cmd)
	}
	return nil, nil
}

func (srv *Server) QueueChange(body io.Reader, form url.Values, ps httprouter.Params) (interface{}, error) {
	var plc PlaylistChange
	if err := json.NewDecoder(body).Decode(&plc); err != nil {
		return nil, err
	}
	srv.ch <- cmdQueueChange(plc)
	return nil, nil
}

func (srv *Server) PlaylistChange(body io.Reader, form url.Values, ps httprouter.Params) (interface{}, error) {
	var plc PlaylistChange
	if err := json.NewDecoder(body).Decode(&plc); err != nil {
		return nil, err
	}
	srv.ch <- cmdPlaylistChange{
		plc:  plc,
		name: ps.ByName("playlist"),
	}
	return nil, nil
}

func (srv *Server) ProtocolRefresh(body io.Reader, form url.Values, ps httprouter.Params) (interface{}, error) {
	var pd ProtocolData
	if err := json.NewDecoder(body).Decode(&pd); err != nil {
		return nil, err
	}
	ch := make(chan error)
	srv.ch <- cmdProtocolRefresh{
		protocol: pd.Protocol,
		key:      pd.Key,
		list:     false,
		doDelete: true,
		err:      ch,
	}
	return nil, <-ch
}

func (srv *Server) ProtocolAdd(body io.Reader, form url.Values, ps httprouter.Params) (interface{}, error) {
	var ap struct {
		Protocol string
		Params   []string
	}
	if err := json.NewDecoder(body).Decode(&ap); err != nil {
		return nil, err
	}
	prot, err := protocol.ByName(ap.Protocol)
	if err != nil {
		return nil, err
	}
	inst, err := prot.NewInstance(ap.Params, nil)
	if err != nil {
		return nil, err
	}
	srv.ch <- cmdProtocolAdd{
		Name:     ap.Protocol,
		Instance: inst,
	}
	return nil, nil
}

func (srv *Server) ProtocolRemove(body io.Reader, form url.Values, ps httprouter.Params) (interface{}, error) {
	var pd ProtocolData
	if err := json.NewDecoder(body).Decode(&pd); err != nil {
		return nil, err
	}
	srv.ch <- cmdProtocolRemove{
		protocol: pd.Protocol,
		key:      pd.Key,
	}
	return nil, nil
}

type ProtocolData struct {
	Protocol string
	Key      string
}

func (srv *Server) TokenRegister(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	srv.ch <- cmdTokenRegister(r.FormValue("token"))
	http.Redirect(w, r, "/", 302)
}
