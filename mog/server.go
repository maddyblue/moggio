// Package mog implements the mog protocol.
package mog

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
	Id   int
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
		Id:       s.Id,
	})
}

type Playlist struct {
	Id    int
	Songs []*Song
}

type Server struct {
	Addr string // TCP address to listen on, ":6601"
	Root string // Root music directory

	Songs    []*Song
	State    State
	Playlist Playlist
	Song     *Song
	Volume   int
	NextSong int
	Elapsed  time.Duration
	Error    string

	nextid int
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
	http.Handle("/", r)

	log.Println("mog: listening on", addr)
	log.Println("mog: Music root:", srv.Root)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) List(w http.ResponseWriter, r *http.Request) {
	t := List(s.Songs)
	b, err := json.Marshal(t)
	if err != nil {
		serveError(w, err)
		return
	}
	w.Write(b)
	return
}

type List []*Song

func (s *Server) Status(w http.ResponseWriter, r *http.Request) {
	t := Status{
		Volume:   s.Volume,
		Playlist: s.Playlist.Id,
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
	return
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
	var songs []*Song
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
					songs = append(songs, &Song{
						Song: s,
						File: p,
						Id:   srv.nextid,
					})
					srv.nextid++
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
