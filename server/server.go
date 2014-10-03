// Package server implements the mog protocol.
package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"code.google.com/p/go.net/websocket"

	"github.com/BurntSushi/toml"
	"github.com/julienschmidt/httprouter"
	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/output"
	"github.com/mjibson/mog/protocol"
)

func ListenAndServe(stateFile, addr string) error {
	server, err := New(stateFile)
	if err != nil {
		return err
	}
	return server.ListenAndServe(addr)
}

const (
	statePlay State = iota
	stateStop
	statePause
)

type State int

func (s State) String() string {
	switch s {
	case statePlay:
		return "play"
	case stateStop:
		return "stop"
	case statePause:
		return "pause"
	}
	return ""
}

type Playlist []SongID

type SongID struct {
	Protocol string
	ID       string
}

func ParseSongID(s string) (id SongID, err error) {
	sp := strings.SplitN(s, "|", 2)
	if len(sp) != 2 {
		return id, fmt.Errorf("bad songid: %v", s)
	}
	return SongID{sp[0], sp[1]}, nil
}

func (s SongID) String() string {
	return fmt.Sprintf("%s|%s", s.Protocol, s.ID)
}

func (s SongID) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
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
	Playlist   Playlist
	PlaylistID int
	Repeat     bool
	Random     bool
	Protocols  map[string]*protocol.Instance

	// Current song data.
	playlistIndex int
	songID        SongID
	song          codec.Song
	info          codec.SongInfo
	elapsed       time.Duration

	ch          chan command
	waitch      chan struct{}
	lock        sync.Locker
	state       State
	songs       map[SongID]codec.Song
	stateFile   string
	savePending bool
}

func (srv *Server) wait() {
	srv.lock.Lock()
	if srv.waitch == nil {
		srv.waitch = make(chan struct{})
	}
	srv.lock.Unlock()
	<-srv.waitch
}

func (srv *Server) broadcast() {
	srv.lock.Lock()
	defer srv.lock.Unlock()
	if srv.waitch == nil {
		return
	}
	close(srv.waitch)
	srv.waitch = nil
}

var dir = filepath.Join("server")

func New(stateFile string) (*Server, error) {
	srv := Server{
		ch:        make(chan command),
		lock:      new(sync.Mutex),
		songs:     make(map[SongID]codec.Song),
		Protocols: make(map[string]*protocol.Instance),
	}
	srv.lock.Lock()
	defer srv.lock.Unlock()
	if stateFile != "" {
		if f, err := os.Open(stateFile); os.IsNotExist(err) {
		} else if err != nil {
			return nil, err
		} else {
			defer f.Close()
			if _, err := toml.DecodeReader(f, &srv); err != nil {
				return nil, err
			}
			for name := range srv.Protocols {
				go func(name string) {
					_, err := srv.ProtocolRefresh(url.Values{"protocol": []string{name}}, nil)
					if err != nil {
						log.Println(err)
					}
				}(name)
			}
		}
		srv.stateFile = stateFile
	}
	for k := range protocol.Get() {
		if srv.Protocols[k] == nil {
			srv.Protocols[k] = new(protocol.Instance)
		}
	}
	go srv.audio()
	go func() {
		for _ = range time.Tick(time.Millisecond * 250) {
			srv.broadcast()
		}
	}()
	return &srv, nil
}

func (s *Server) Save() {
	if s.stateFile == "" {
		return
	}
	go func() {
		s.lock.Lock()
		defer s.lock.Unlock()
		if s.savePending {
			return
		}
		s.savePending = true
		time.AfterFunc(time.Second, s.save)
	}()
}

func (s *Server) save() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.savePending = false
	tmp := s.stateFile + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		log.Println(err)
		return
	}
	if err := toml.NewEncoder(f).Encode(s); err != nil {
		log.Println(err)
		return
	}
	if err := os.Rename(tmp, s.stateFile); err != nil {
		log.Println(err)
		return
	}
}

func (srv *Server) GetMux() *http.ServeMux {
	router := httprouter.New()
	router.GET("/api/status", JSON(srv.Status))
	router.GET("/api/list", JSON(srv.List))
	router.GET("/api/playlist/change", JSON(srv.PlaylistChange))
	router.GET("/api/playlist/get", JSON(srv.PlaylistGet))
	router.GET("/api/protocol/update", JSON(srv.ProtocolUpdate))
	router.GET("/api/protocol/get", JSON(srv.ProtocolGet))
	router.GET("/api/protocol/list", JSON(srv.ProtocolList))
	router.GET("/api/protocol/refresh", JSON(srv.ProtocolRefresh))
	router.GET("/api/song/info", JSON(srv.SongInfo))
	router.GET("/api/cmd/:cmd", JSON(srv.Cmd))
	router.GET("/api/oauth/:protocol", srv.OAuth)
	fs := http.FileServer(http.Dir(dir))
	mux := http.NewServeMux()
	mux.Handle("/static/", fs)
	mux.HandleFunc("/", index)
	mux.Handle("/api/", router)
	mux.Handle("/ws/", websocket.Handler(srv.WebSocket))
	return mux
}

// ListenAndServe listens on the TCP network address addr and then calls
// Serve to handle requests on incoming connections.
func (srv *Server) ListenAndServe(addr string) error {
	mux := srv.GetMux()
	log.Println("mog: listening on", addr)
	return http.ListenAndServe(addr, mux)
}

func (srv *Server) WebSocket(ws *websocket.Conn) {
	for {
		srv.wait()
		if err := websocket.JSON.Send(ws, srv.status()); err != nil {
			log.Println(err)
			break
		}
	}
}

func index(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, filepath.Join(dir, "static", "index.html"))
}

func (srv *Server) audio() {
	var o output.Output
	var t chan interface{}
	var present bool
	var dur time.Duration
	srv.state = stateStop
	var next, stop, tick, play, pause, prev func()
	prev = func() {
		log.Println("prev")
		srv.playlistIndex--
		if srv.elapsed < time.Second*3 {
			srv.playlistIndex--
		}
		next()
	}
	pause = func() {
		log.Println("pause")
		switch srv.state {
		case statePause, stateStop:
			log.Println("pause: resume")
			t = make(chan interface{})
			close(t)
			tick()
			srv.state = statePlay
		case statePlay:
			log.Println("pause: pause")
			t = nil
			srv.state = statePause
		}
	}
	next = func() {
		log.Println("next")
		stop()
		play()
	}
	stop = func() {
		log.Println("stop")
		srv.state = stateStop
		t = nil
		srv.song = nil
	}
	tick = func() {
		if false && srv.elapsed > srv.info.Time {
			log.Println("elapsed time completed", srv.elapsed, srv.info.Time)
			stop()
		}
		if srv.song == nil {
			if len(srv.Playlist) == 0 {
				log.Println("empty playlist")
				stop()
				return
			} else if srv.playlistIndex >= len(srv.Playlist) {
				if srv.Repeat {
					srv.playlistIndex = 0
				} else {
					log.Println("end of playlist")
					stop()
					return
				}
			}
			srv.songID = srv.Playlist[srv.playlistIndex]
			srv.song, present = srv.songs[srv.songID]
			srv.playlistIndex++
			if !present {
				return
			}
			sr, ch, err := srv.song.Init()
			if err != nil {
				panic(err)
			}
			o, err = output.Get(sr, ch)
			if err != nil {
				panic(fmt.Errorf("mog: could not open audio (%v, %v): %v", sr, ch, err))
			}
			srv.info, err = srv.song.Info()
			if err != nil {
				log.Println(err)
				next()
				return
			}
			srv.elapsed = 0
			dur = time.Second / (time.Duration(sr * ch))
			log.Println("playing", srv.info, sr, ch, dur, time.Duration(4096)*dur)
			t = make(chan interface{})
			close(t)
			srv.state = statePlay
		}
		const expected = 4096
		next, err := srv.song.Play(expected)
		if err == nil {
			srv.elapsed += time.Duration(len(next)) * dur
			if len(next) > 0 {
				o.Push(next)
			}
		}
		if len(next) < expected || err != nil {
			log.Println("end of song", len(next), expected, err)
			stop()
		}
	}
	play = func() {
		log.Println("play")
		if srv.playlistIndex > len(srv.Playlist) {
			srv.playlistIndex = 0
		}
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
			case cmdNext:
				next()
			case cmdPause:
				pause()
			case cmdPrev:
				prev()
			default:
				panic("unknown command")
			}
		}
	}
}

type command int

const (
	cmdPlay command = iota
	cmdStop
	cmdNext
	cmdPause
	cmdPrev
)

func JSON(h func(url.Values, httprouter.Params) (interface{}, error)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if err := r.ParseForm(); err != nil {
			serveError(w, err)
			return
		}
		d, err := h(r.Form, ps)
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
	name := ps.ByName("protocol")
	prot, err := protocol.ByName(name)
	if err != nil {
		serveError(w, err)
		return
	}
	instance, ok := srv.Protocols[name]
	if !ok || prot.OAuth == nil {
		serveError(w, fmt.Errorf("bad protocol"))
		return
	}
	t, err := prot.OAuth.Exchange(r.FormValue("code"))
	if err != nil {
		serveError(w, err)
		return
	}
	instance.OAuthToken = t
	// "Bearer" was added for dropbox. It happens to work also with Google Music's
	// OAuth. This may need to be changed to be protocol-specific in the future.
	instance.OAuthToken.TokenType = "Bearer"
	srv.Save()
	go srv.ProtocolRefresh(url.Values{"protocol": []string{name}}, nil)
	http.Redirect(w, r, "/", http.StatusFound)
}

func (srv *Server) Cmd(form url.Values, ps httprouter.Params) (interface{}, error) {
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
	default:
		return nil, fmt.Errorf("unknown command: %v", cmd)
	}
	return nil, nil
}

func (srv *Server) SongInfo(form url.Values, ps httprouter.Params) (interface{}, error) {
	var si []codec.SongInfo
	for _, s := range form["song"] {
		id, err := ParseSongID(s)
		if err != nil {
			return nil, err
		}
		song, ok := srv.songs[id]
		if !ok {
			return nil, fmt.Errorf("unknown song: %v", id)
		}
		info, err := song.Info()
		if err != nil {
			return nil, err
		}
		si = append(si, info)
	}
	return si, nil
}

func (srv *Server) PlaylistGet(form url.Values, ps httprouter.Params) (interface{}, error) {
	return srv.Playlist, nil
}

func (srv *Server) ProtocolGet(form url.Values, ps httprouter.Params) (interface{}, error) {
	return protocol.Get(), nil
}

func (srv *Server) ProtocolList(form url.Values, ps httprouter.Params) (interface{}, error) {
	return srv.Protocols, nil
}

func (srv *Server) ProtocolRefresh(form url.Values, ps httprouter.Params) (interface{}, error) {
	p := form.Get("protocol")
	inst := srv.Protocols[p]
	if inst == nil {
		return nil, fmt.Errorf("unknown protocol")
	}
	if len(inst.Params) == 0 && inst.OAuthToken == nil {
		return nil, nil
	}
	// On error, still process songs. Needed because it is valid to clear
	// instances.
	songs, err := protocol.ListSongs(p, inst)
	for id := range srv.songs {
		if id.Protocol == p {
			delete(srv.songs, id)
		}
	}
	for id, s := range songs {
		srv.songs[SongID{Protocol: p, ID: id}] = s
	}
	srv.Save()
	return nil, err
}

func (srv *Server) ProtocolUpdate(form url.Values, ps httprouter.Params) (interface{}, error) {
	p := form.Get("protocol")
	params := form["params"]
	srv.Protocols[p].Params = params
	return srv.ProtocolRefresh(form, ps)
}

// Takes form values:
// * clear: if set to anything will clear playlist
// * remove/add: song ids
// Duplicate songs will not be added.
func (srv *Server) PlaylistChange(form url.Values, ps httprouter.Params) (interface{}, error) {
	srv.PlaylistID++
	srv.playlistIndex = 0
	t := PlaylistChange{
		PlaylistId: srv.PlaylistID,
	}
	if len(form["clear"]) > 0 {
		srv.Playlist = nil
		srv.ch <- cmdStop
	}
	m := make(map[SongID]int)
	for i, id := range srv.Playlist {
		m[id] = i
	}
	for _, rem := range form["remove"] {
		sp := strings.SplitN(rem, "|", 2)
		if len(sp) != 2 {
			t.Error("bad id: %v", rem)
			continue
		}
		id := SongID{sp[0], sp[1]}
		if s, ok := srv.songs[id]; !ok {
			t.Error("unknown id: %v", rem)
		} else if s == srv.song {
			srv.ch <- cmdStop
		}
		delete(m, id)
	}
	for _, add := range form["add"] {
		sp := strings.SplitN(add, "|", 2)
		if len(sp) != 2 {
			t.Error("bad id: %v", add)
			continue
		}
		id := SongID{sp[0], sp[1]}
		if _, ok := srv.songs[id]; !ok {
			t.Error("unknown id: %v", add)
		}
		m[id] = len(m)
	}
	srv.Playlist = make(Playlist, len(m))
	for songid, index := range m {
		srv.Playlist[index] = songid
	}
	srv.Save()
	return &t, nil
}

type PlaylistChange struct {
	PlaylistId int
	Errors     []string
}

func (p *PlaylistChange) Error(format string, a ...interface{}) {
	p.Errors = append(p.Errors, fmt.Sprintf(format, a...))
}

func (s *Server) List(form url.Values, ps httprouter.Params) (interface{}, error) {
	songs := make([]SongID, 0)
	for id := range s.songs {
		songs = append(songs, id)
	}
	return songs, nil
}

func (s *Server) status() *Status {
	return &Status{
		Playlist: s.PlaylistID,
		State:    s.state,
		Song:     s.songID,
		Elapsed:  s.elapsed.Seconds(),
		Time:     s.info.Time.Seconds(),
	}
}

func (s *Server) Status(form url.Values, ps httprouter.Params) (interface{}, error) {
	return s.status(), nil
}

type Status struct {
	// Playlist ID.
	Playlist int
	// Playback state
	State State
	// Song ID.
	Song SongID
	// Elapsed time of current song in seconds.
	Elapsed float64
	// Duration of current song in seconds.
	Time float64
}

func serveError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
