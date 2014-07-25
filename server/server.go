// Package server implements the mog protocol.
package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/output"
	"github.com/mjibson/mog/protocol"
)

const (
	DefaultAddr = ":6601"
)

func ListenAndServe(addr string) error {
	server := &Server{Addr: addr}
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

type Playlist []SongID

type SongID struct {
	Protocol string
	ID       string
}

func (s SongID) String() string {
	return fmt.Sprintf("%s|%s", s.Protocol, s.ID)
}

func (s SongID) MarshalJSON() ([]byte, error) {
	return json.Marshal([2]string{s.Protocol, s.ID})
}

func (s *SongID) UnmarshalJSON(b []byte) error {
	var v [2]string
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	s.Protocol = v[0]
	s.ID = v[1]
	return nil
}

type Server struct {
	Addr string // TCP address to listen on, ":6601"

	Songs      map[SongID]codec.Song
	State      State
	Playlist   Playlist
	PlaylistID int
	// Index of current song in the playlist.
	PlaylistIndex int
	Song          codec.Song
	Info          codec.SongInfo
	Volume        int
	Elapsed       time.Duration
	Error         string
	Repeat        bool
	Random        bool
	Protocols     map[string][]string

	songID int
	ch     chan command
}

var dir = filepath.Join("server")

// ListenAndServe listens on the TCP network address srv.Addr and then calls
// Serve to handle requests on incoming connections. If srv.Addr is blank,
// ":6601" is used.
func (srv *Server) ListenAndServe() error {
	srv.ch = make(chan command)
	srv.Songs = make(map[SongID]codec.Song)
	srv.Protocols = make(map[string][]string)
	go srv.audio()

	addr := srv.Addr
	if addr == "" {
		addr = DefaultAddr
	}
	http.Handle("/api/status", JSON(srv.Status))
	http.Handle("/api/list", JSON(srv.List))
	http.Handle("/api/playlist/change", JSON(srv.PlaylistChange))
	http.Handle("/api/playlist/get", JSON(srv.PlaylistGet))
	http.Handle("/api/protocol/update", JSON(srv.ProtocolUpdate))
	http.Handle("/api/protocol/get", JSON(srv.ProtocolGet))
	http.Handle("/api/protocol/list", JSON(srv.ProtocolList))
	http.Handle("/api/play", JSON(srv.Play))
	fs := http.FileServer(http.Dir(dir))
	http.Handle("/static/", fs)
	http.HandleFunc("/", index)

	log.Println("mog: listening on", addr)
	return http.ListenAndServe(addr, nil)
}

func index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(dir, "static", "index.html"))
}

func (srv *Server) audio() {
	var o output.Output
	var t chan interface{}
	var present bool
	var dur time.Duration
	stop := func() {
		log.Println("stop")
		t = nil
		srv.Song = nil
	}
	tick := func() {
		if srv.Elapsed > srv.Info.Time {
			stop()
		}
		if srv.Song == nil {
			if len(srv.Playlist) == 0 {
				log.Println("empty playlist")
				stop()
				return
			} else if srv.PlaylistIndex >= len(srv.Playlist) {
				if srv.Repeat {
					srv.PlaylistIndex = 0
				} else {
					log.Println("end of playlist")
					stop()
					return
				}
			}
			srv.Song, present = srv.Songs[srv.Playlist[srv.PlaylistIndex]]
			srv.PlaylistIndex++
			if !present {
				return
			}
			sr, ch, err := srv.Song.Init()
			if err != nil {
				log.Fatal(err)
			}
			if o != nil {
				o.Dispose()
			}
			o, err = output.NewPort(sr, ch)
			if err != nil {
				log.Fatal(fmt.Errorf("mog: could not open audio (%v, %v): %v", sr, ch, err))
			}
			srv.Info = srv.Song.Info()
			fmt.Println("playing", srv.Info)
			srv.Elapsed = 0
			dur = time.Second / (time.Duration(sr))
			t = make(chan interface{})
			close(t)
		}
		const expected = 4096
		next, err := srv.Song.Play(expected)
		if err == nil {
			srv.Elapsed += time.Duration(len(next)) * dur
			if len(next) > 0 {
				o.Push(next)
			}
		} else {
			log.Println(err)
		}
		if len(next) < expected || err != nil {
			stop()
		}
	}
	play := func() {
		log.Println("play")
		tick()
	}
	for {
		select {
		case <-t:
			tick()
		case cmd := <-srv.ch:
			switch cmd {
			case cmdPlay:
				play()
			case cmdStop:
				stop()
			default:
				log.Fatal("unknown command")
			}
		}
	}
}

type command int

const (
	cmdPlay command = iota
	cmdStop
)

func JSON(h func(http.ResponseWriter, *http.Request) (interface{}, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		d, err := h(w, r)
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

func (srv *Server) Play(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	srv.ch <- cmdPlay
	return nil, nil
}

func (srv *Server) PlaylistGet(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return srv.Playlist, nil
}

func (srv *Server) ProtocolGet(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return protocol.Get(), nil
}
func (srv *Server) ProtocolList(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	return srv.Protocols, nil
}

func (srv *Server) ProtocolUpdate(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	p := r.FormValue("protocol")
	params := r.Form["params"]
	songs, err := protocol.List(p, params)
	if err != nil {
		return nil, err
	}
	srv.Protocols[p] = params
	for id := range srv.Songs {
		if id.Protocol == p {
			delete(srv.Songs, id)
		}
	}
	for id, s := range songs {
		srv.Songs[SongID{Protocol: p, ID: id}] = s
	}
	return nil, nil
}

// Takes form values:
// * clear: if set to anything will clear playlist
// * remove/add: song ids
// Duplicate songs will not be added.
func (srv *Server) PlaylistChange(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	srv.PlaylistID++
	srv.PlaylistIndex = 0
	t := PlaylistChange{
		PlaylistId: srv.PlaylistID,
	}
	if len(r.Form["clear"]) > 0 {
		srv.Playlist = nil
		srv.ch <- cmdStop
	}
	m := make(map[SongID]int)
	for i, id := range srv.Playlist {
		m[id] = i
	}
	for _, rem := range r.Form["remove"] {
		sp := strings.SplitN(rem, "|", 2)
		if len(sp) != 2 {
			t.Error("bad id: %v", rem)
			continue
		}
		id := SongID{sp[0], sp[1]}
		if s, ok := srv.Songs[id]; !ok {
			t.Error("unknown id: %v", rem)
		} else if s == srv.Song {
			srv.ch <- cmdStop
		}
		delete(m, id)
	}
	for _, add := range r.Form["add"] {
		sp := strings.SplitN(add, "|", 2)
		if len(sp) != 2 {
			t.Error("bad id: %v", add)
			continue
		}
		id := SongID{sp[0], sp[1]}
		if _, ok := srv.Songs[id]; !ok {
			t.Error("unknown id: %v", add)
		}
		m[id] = len(m)
	}
	srv.Playlist = make(Playlist, len(m))
	for songid, index := range m {
		srv.Playlist[index] = songid
	}
	return &t, nil
}

type PlaylistChange struct {
	PlaylistId int
	Errors     []string
}

func (p *PlaylistChange) Error(format string, a ...interface{}) {
	p.Errors = append(p.Errors, fmt.Sprintf(format, a...))
}

func (s *Server) List(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	songs := make([]SongID, 0)
	for id := range s.Songs {
		songs = append(songs, id)
	}
	return songs, nil
}

func (s *Server) Status(w http.ResponseWriter, r *http.Request) (interface{}, error) {
	t := Status{
		Volume:   s.Volume,
		Playlist: s.PlaylistID,
		State:    s.State,
		//Song:     s.Song.Id,
		Elapsed: s.Elapsed,
	}
	return &t, nil
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

func serveError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
