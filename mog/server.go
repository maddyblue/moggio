// Package mog implements the mog protocol.
package mog

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/mjibson/mog/codec"
)

const (
	// Version is the current protocol version.
	Version     = "0.0.0"
	DefaultAddr = ":6601"
)

func ListenAndServe(addr, root string) error {
	server := &Server{Addr: addr, Root: root}
	return server.ListenAndServe()
}

const (
	STATE_PLAY State = iota
	STATE_STOP
	STATE_PAUSE
)

type State int

func (s State) String() string {
	switch s {
	case STATE_PLAY:
		return "play"
	case STATE_STOP:
		return "stop"
	case STATE_PAUSE:
		return "pause"
	}
	return ""
}

type Song struct {
	codec.Song
	File string
}

func (s *Song) MarshalJSON() ([]byte, error) {
	type S struct {
		codec.SongInfo
		File string
		Id   int
	}
	return json.Marshal(&S{
		SongInfo: s.Info(),
		File:     s.File,
	})
}

type Playlist []int

type Server struct {
	Addr string // TCP address to listen on, ":6601"
	Root string // Root music directory

	Songs      Songs
	State      State
	Playlist   Playlist
	PlaylistID int
	Song       *Song
	Volume     int
	NextSong   int
	Elapsed    time.Duration
	Error      string

	songID int
}

// ListenAndServe listens on the TCP network address srv.Addr and then calls
// Serve to handle requests on incoming connections. If srv.Addr is blank,
// ":6601" is used.
func (srv *Server) ListenAndServe() error {
	f, e := os.Open(srv.Root)
	if e != nil {
		return e
	}
	fi, e := f.Stat()
	if e != nil {
		return e
	}
	if !fi.IsDir() {
		return fmt.Errorf("mog: not a directory: %s", srv.Root)
	}
	srv.Update()

	addr := srv.Addr
	if addr == "" {
		addr = DefaultAddr
	}
	r := mux.NewRouter()
	r.HandleFunc("/status", srv.Status)
	r.HandleFunc("/list", srv.List)
	r.HandleFunc("/playlist/change", srv.PlaylistChange)
	r.HandleFunc("/playlist/get", srv.PlaylistGet)
	http.Handle("/", r)

	log.Println("mog: listening on", addr)
	log.Println("mog: Music root:", srv.Root)
	return http.ListenAndServe(addr, nil)
}

func (srv *Server) PlaylistGet(w http.ResponseWriter, r *http.Request) {
	b, err := json.Marshal(srv.Playlist)
	if err != nil {
		serveError(w, err)
		return
	}
	w.Write(b)
}

// Takes form values:
// * clear: if set to anything will clear playlist
// * remove/add: song ids
// Duplicate songs will not be added.
func (srv *Server) PlaylistChange(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		serveError(w, err)
		return
	}
	srv.PlaylistID++
	t := PlaylistChange{
		PlaylistId: srv.PlaylistID,
	}
	if len(r.Form["clear"]) > 0 {
		srv.Playlist = nil
	}
	// songid -> index
	m := make(map[int]int)
	for i, id := range srv.Playlist {
		m[id] = i
	}
	for _, id := range r.Form["remove"] {
		i, err := strconv.Atoi(id)
		if err != nil {
			log.Println("mog:", err)
			continue
		}
		if _, ok := srv.Songs[i]; !ok {
			log.Println("mog: unknown song id:", i)
			continue
		}
		if idx, present := m[i]; present {
			srv.Playlist = append(srv.Playlist[:idx], srv.Playlist[idx+1:]...)
			delete(m, i)
			t.Removed = append(t.Removed, i)
		}
	}
	for _, id := range r.Form["add"] {
		i, err := strconv.Atoi(id)
		if err != nil {
			log.Println("mog:", err)
			continue
		}
		if _, ok := srv.Songs[i]; !ok {
			log.Println("mog: unknown song id:", i)
			continue
		}
		if _, present := m[i]; !present {
			srv.Playlist = append(srv.Playlist, i)
			m[i] = len(srv.Playlist)
			t.Added = append(t.Added, i)
		}
	}
	b, err := json.Marshal(&t)
	if err != nil {
		serveError(w, err)
		return
	}
	w.Write(b)
}

type PlaylistChange struct {
	PlaylistId int
	Added      []int
	Removed    []int
}

func (s *Server) List(w http.ResponseWriter, r *http.Request) {
	t := Songs(s.Songs)
	b, err := json.Marshal(&t)
	if err != nil {
		serveError(w, err)
		return
	}
	w.Write(b)
}

type Songs map[int]*Song
type _Songs map[string]*Song

func (s Songs) MarshalJSON() ([]byte, error) {
	m := make(_Songs)
	for k, v := range s {
		m[strconv.Itoa(k)] = v
	}
	return json.Marshal(&m)
}

func (s Songs) UnmarshalJSON(b []byte) error {
	var _s _Songs
	if err := json.Unmarshal(b, &_s); err != nil {
		return err
	}
	for k, v := range _s {
		i, err := strconv.Atoi(k)
		if err != nil {
			return err
		}
		s[i] = v
	}
	return nil
}

func (s *Server) Status(w http.ResponseWriter, r *http.Request) {
	t := Status{
		Volume:   s.Volume,
		Playlist: s.PlaylistID,
		State:    s.State,
		//Song:     s.Song.Id,
		Elapsed: s.Elapsed,
	}
	b, err := json.Marshal(&t)
	if err != nil {
		serveError(w, err)
		return
	}
	w.Write(b)
}

type Status struct {
	// Volume from 0 - 100.
	Volume int
	// Playlist ID.
	Playlist int
	// Playback state
	State State
	// Song ID.
	Song int
	// Elapsed time of current song.
	Elapsed time.Duration
	// Duration of current song.
	Time time.Duration
}

func (srv *Server) Update() {
	songs := make(Songs)
	var walk func(string)
	walk = func(dirname string) {
		f, err := os.Open(dirname)
		if err != nil {
			return
		}
		fis, err := f.Readdir(0)
		if err != nil {
			return
		}
		for _, fi := range fis {
			p := filepath.Join(dirname, fi.Name())
			if fi.IsDir() {
				walk(p)
			} else {
				f, err := os.Open(p)
				if err != nil {
					continue
				}
				ss, _, err := codec.Decode(f)
				if err != nil {
					continue
				}
				for _, s := range ss {
					songs[srv.songID] = &Song{
						Song: s,
						File: p,
					}
					srv.songID++
				}
			}
		}
	}
	walk(srv.Root)
	srv.Songs = songs
}

func serveError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
