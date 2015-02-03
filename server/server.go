// Package server implements the mog protocol.
package server

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/output"
	"github.com/mjibson/mog/protocol"
	"golang.org/x/net/websocket"
	"golang.org/x/oauth2"
)

func printErr(e error) {
	log.Println(e)
	b := make([]byte, 4096)
	runtime.Stack(b, false)
	fmt.Println(string(b))
}

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
	Key      string
	ID       string
}

func ParseSongID(s string) (id SongID, err error) {
	sp := strings.SplitN(s, "|", 3)
	if len(sp) != 3 {
		return id, fmt.Errorf("bad songid: %v", s)
	}
	return SongID{sp[0], sp[1], sp[2]}, nil
}

func (s SongID) String() string {
	return fmt.Sprintf("%s|%s|%s", s.Protocol, s.Key, s.ID)
}

func (s SongID) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Protocol string
		Key      string
		ID       string
		UID      string
	}{
		Protocol: s.Protocol,
		Key:      s.Key,
		ID:       s.ID,
		UID:      s.String(),
	})
}

type Server struct {
	lock sync.Mutex

	Queue     Playlist
	Playlists map[string]Playlist

	Repeat    bool
	Random    bool
	Protocols map[string]map[string]protocol.Instance

	// Current song data.
	PlaylistIndex int
	songID        SongID
	song          codec.Song
	info          codec.SongInfo
	elapsed       time.Duration

	ch          chan command
	chParam     chan interface{}
	waiters     map[chan *waitData]struct{}
	state       State
	songs       map[SongID]*codec.SongInfo
	stateFile   string
	savePending bool
}

type waitData struct {
	Type waitType
	Data interface{}
}

type waitType string

const (
	waitStatus    waitType = "status"
	waitPlaylist           = "playlist"
	waitProtocols          = "protocols"
	waitTracks             = "tracks"
)

func (srv *Server) makeWaitData(wt waitType) *waitData {
	var data interface{}
	switch wt {
	case waitProtocols:
		protos := make(map[string][]string)
		for p, m := range srv.Protocols {
			for key := range m {
				protos[p] = append(protos[p], key)
			}
		}
		data = struct {
			Available map[string]protocol.Params
			Current   map[string][]string
		}{
			protocol.Get(),
			protos,
		}
	case waitStatus:
		data = srv.status()
	case waitTracks:
		data = struct {
			Tracks []listItem
		}{
			Tracks: srv.list(),
		}
	case waitPlaylist:
		d := struct {
			Queue     PlaylistInfo
			Playlists map[string]PlaylistInfo
		}{
			Queue:     srv.playlistInfo(srv.Queue),
			Playlists: make(map[string]PlaylistInfo),
		}
		for name, p := range srv.Playlists {
			d.Playlists[name] = srv.playlistInfo(p)
		}
		data = d
	default:
		panic("bad wait type")
	}
	return &waitData{
		Type: wt,
		Data: data,
	}
}

type PlaylistInfo []listItem

func (srv *Server) playlistInfo(p Playlist) PlaylistInfo {
	r := make(PlaylistInfo, len(p))
	for idx, id := range p {
		r[idx] = listItem{
			ID: id,
			Info: srv.songs[id],
		}
	}
	return r
}

func (srv *Server) broadcast(wt waitType) {
	wd := srv.makeWaitData(wt)
	srv.lock.Lock()
	defer srv.lock.Unlock()
	for w := range srv.waiters {
		go func(w chan *waitData) {
			w <- wd
		}(w)
	}
}

var dir = filepath.Join("server")

func New(stateFile string) (*Server, error) {
	srv := Server{
		ch:        make(chan command),
		chParam:   make(chan interface{}, 1),
		songs:     make(map[SongID]*codec.SongInfo),
		Protocols: make(map[string]map[string]protocol.Instance),
		Playlists: make(map[string]Playlist),
		waiters:   make(map[chan *waitData]struct{}),
	}
	for name := range protocol.Get() {
		srv.Protocols[name] = make(map[string]protocol.Instance)
	}
	srv.lock.Lock()
	defer srv.lock.Unlock()
	if stateFile != "" {
		if f, err := os.Open(stateFile); os.IsNotExist(err) {
		} else if err != nil {
			return nil, err
		} else {
			defer f.Close()
			if err := gob.NewDecoder(f).Decode(&srv); err != nil {
				return nil, err
			}
			for name, insts := range srv.Protocols {
				for key := range insts {
					go func(name, key string) {
						if err := srv.protocolRefresh(name, key, true); err != nil {
							log.Println(err)
						}
					}(name, key)
				}
			}
		}
		srv.stateFile = stateFile
	}
	go srv.audio()
	return &srv, nil
}

func (srv *Server) Save() {
	if srv.stateFile == "" {
		return
	}
	go func() {
		srv.lock.Lock()
		defer srv.lock.Unlock()
		if srv.savePending {
			return
		}
		srv.savePending = true
		time.AfterFunc(time.Second, srv.save)
	}()
}

func (srv *Server) save() {
	srv.lock.Lock()
	defer srv.lock.Unlock()
	srv.savePending = false
	tmp := srv.stateFile + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		log.Println(err)
		return
	}
	if err := gob.NewEncoder(f).Encode(srv); err != nil {
		log.Println(err)
		return
	}
	f.Close()
	if err := os.Rename(tmp, srv.stateFile); err != nil {
		log.Println(err)
		return
	}
}

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
	indexHTML, err = ioutil.ReadAll(index)
	if err != nil {
		log.Fatal(err)
	}
	router := httprouter.New()
	router.GET("/api/cmd/:cmd", JSON(srv.Cmd))
	router.GET("/api/oauth/:protocol", srv.OAuth)
	router.POST("/api/cmd/:cmd", JSON(srv.Cmd))
	router.POST("/api/queue/change", JSON(srv.QueueChange))
	router.POST("/api/playlist/change/:playlist", JSON(srv.PlaylistChange))
	router.POST("/api/protocol/add", JSON(srv.ProtocolAdd))
	router.POST("/api/protocol/remove", JSON(srv.ProtocolRemove))
	router.POST("/api/protocol/refresh", JSON(srv.ProtocolRefresh))
	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(webFS))
	mux.HandleFunc("/", Index)
	mux.Handle("/api/", router)
	mux.Handle("/ws/", websocket.Handler(srv.WebSocket))
	return mux
}

// ListenAndServe listens on the TCP network address addr and then calls
// Serve to handle requests on incoming connections.
func (srv *Server) ListenAndServe(addr string) error {
	mux := srv.GetMux(true)
	log.Println("mog: listening on", addr)
	return http.ListenAndServe(addr, mux)
}

func (srv *Server) WebSocket(ws *websocket.Conn) {
	inits := []waitType{
		waitPlaylist,
		waitProtocols,
		waitStatus,
		waitTracks,
	}
	for _, wt := range inits {
		if err := websocket.JSON.Send(ws, srv.makeWaitData(wt)); err != nil {
			log.Println(err)
			return
		}
	}
	c := make(chan *waitData)
	srv.lock.Lock()
	srv.waiters[c] = struct{}{}
	srv.lock.Unlock()
	for d := range c {
		go func(d *waitData) {
			if err := websocket.JSON.Send(ws, d); err != nil {
				srv.lock.Lock()
				if _, ok := srv.waiters[c]; ok {
					delete(srv.waiters, c)
					close(c)
				}
				srv.lock.Unlock()
			}
		}(d)
	}
}

func Index(w http.ResponseWriter, r *http.Request) {
	w.Write(indexHTML)
}

func (srv *Server) audio() {
	var o output.Output
	var t chan interface{}
	var dur time.Duration
	srv.state = stateStop
	var next, stop, tick, play, pause, prev func()
	var timer <-chan time.Time
	prev = func() {
		log.Println("prev")
		srv.PlaylistIndex--
		if srv.elapsed < time.Second*3 {
			srv.PlaylistIndex--
		}
		if srv.PlaylistIndex < 0 {
			srv.PlaylistIndex = 0
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
		if srv.song != nil {
			if srv.Random && len(srv.Queue) > 1 {
				n := srv.PlaylistIndex
				for n == srv.PlaylistIndex {
					n = rand.Intn(len(srv.Queue))
				}
				srv.PlaylistIndex = n
			} else {
				srv.PlaylistIndex++
			}
		}
		srv.song = nil
	}
	tick = func() {
		if false && srv.elapsed > srv.info.Time {
			log.Println("elapsed time completed", srv.elapsed, srv.info.Time)
			stop()
		}
		if srv.song == nil {
			if len(srv.Queue) == 0 {
				log.Println("empty queue")
				stop()
				return
			}
			if srv.PlaylistIndex >= len(srv.Queue) {
				if srv.Repeat {
					srv.PlaylistIndex = 0
				} else {
					log.Println("end of queue", srv.PlaylistIndex, len(srv.Queue))
					stop()
					return
				}
			}

			srv.songID = srv.Queue[srv.PlaylistIndex]
			sid := srv.songID
			song, err := srv.Protocols[sid.Protocol][sid.Key].GetSong(sid.ID)
			if err != nil {
				printErr(err)
				next()
				return
			}
			srv.song = song
			sr, ch, err := srv.song.Init()
			if err != nil {
				srv.song.Close()
				printErr(err)
				next()
				return
			}
			o, err = output.Get(sr, ch)
			if err != nil {
				printErr(fmt.Errorf("mog: could not open audio (%v, %v): %v", sr, ch, err))
				next()
				return
			}
			srv.info = *srv.songs[sid]
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
			select {
			case <-timer:
				srv.broadcast(waitStatus)
				timer = nil
			default:
			}
			if timer == nil {
				timer = time.After(time.Millisecond * 500)
			}
		}
		if len(next) < expected || err != nil {
			log.Println("end of song", len(next), expected, err)
			if err == io.ErrUnexpectedEOF {
				log.Println("attempting to restart song")
				n := srv.PlaylistIndex
				stop()
				srv.PlaylistIndex = n
				play()
			} else {
				stop()
				play()
			}
		}
	}
	play = func() {
		log.Println("play")
		if srv.PlaylistIndex > len(srv.Queue) {
			srv.PlaylistIndex = 0
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
			case cmdPlayIdx:
				stop()
				srv.PlaylistIndex = (<-srv.chParam).(int)
				play()
			default:
				panic("unknown command")
			}
			srv.broadcast(waitStatus)
			srv.Save()
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
	cmdPlayIdx
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
	prots, ok := srv.Protocols[name]
	if !ok || prot.OAuth == nil {
		serveError(w, fmt.Errorf("bad protocol"))
		return
	}
	t, err := prot.OAuth.Exchange(oauth2.NoContext, r.FormValue("code"))
	if err != nil {
		serveError(w, err)
		return
	}
	// "Bearer" was added for dropbox. It happens to work also with Google Music's
	// OAuth. This may need to be changed to be protocol-specific in the future.
	t.TokenType = "Bearer"
	instance, err := prot.NewInstance(nil, t)
	if err != nil {
		serveError(w, err)
		return
	}
	prots[t.AccessToken] = instance
	srv.Save()
	go srv.protocolRefresh(name, instance.Key(), false)
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
	case "play_idx":
		i, err := strconv.Atoi(form.Get("idx"))
		if err != nil {
			return nil, err
		}
		srv.chParam <- i
		srv.ch <- cmdPlayIdx
	case "random":
		srv.lock.Lock()
		srv.Random = !srv.Random
		srv.lock.Unlock()
		srv.Save()
		srv.broadcast(waitStatus)
	case "repeat":
		srv.lock.Lock()
		srv.Repeat = !srv.Repeat
		srv.lock.Unlock()
		srv.Save()
		srv.broadcast(waitStatus)
	default:
		return nil, fmt.Errorf("unknown command: %v", cmd)
	}
	return nil, nil
}

func (srv *Server) GetInstance(name, key string) (protocol.Instance, error) {
	prots, ok := srv.Protocols[name]
	if !ok {
		return nil, fmt.Errorf("unknown protocol: %s", name)
	}
	inst := prots[key]
	if inst == nil {
		return nil, fmt.Errorf("unknown key: %s", key)
	}
	return inst, nil
}

func (srv *Server) protocolRefresh(protocol, key string, list bool) error {
	inst, err := srv.GetInstance(protocol, key)
	if err != nil {
		return err
	}
	f := inst.Refresh
	if list {
		f = inst.List
	}
	songs, err := f()
	if err != nil {
		return err
	}
	srv.lock.Lock()
	for id := range srv.songs {
		if id.Protocol == protocol {
			delete(srv.songs, id)
		}
	}
	for id, s := range songs {
		srv.songs[SongID{
			Protocol: protocol,
			Key:      key,
			ID:       id,
		}] = s
	}
	srv.lock.Unlock()
	srv.Save()
	srv.broadcast(waitTracks)
	srv.broadcast(waitProtocols)
	return err
}

func (srv *Server) ProtocolRefresh(form url.Values, ps httprouter.Params) (interface{}, error) {
	return nil, srv.protocolRefresh(form.Get("protocol"), form.Get("key"), false)
}

func (srv *Server) ProtocolAdd(form url.Values, ps httprouter.Params) (interface{}, error) {
	p := form.Get("protocol")
	prot, err := protocol.ByName(p)
	if err != nil {
		return nil, err
	}
	inst, err := prot.NewInstance(form["params"], nil)
	if err != nil {
		return nil, err
	}
	srv.Protocols[p][inst.Key()] = inst
	err = srv.protocolRefresh(p, inst.Key(), false)
	if err != nil {
		delete(srv.Protocols[p], inst.Key())
		return nil, err
	}
	return nil, nil
}

func (srv *Server) ProtocolRemove(form url.Values, ps httprouter.Params) (interface{}, error) {
	p := form.Get("protocol")
	k := form.Get("key")
	srv.lock.Lock()
	prots, ok := srv.Protocols[p]
	if !ok {
		return nil, fmt.Errorf("unknown protocol: %v", p)
	}
	delete(prots, k)
	for id := range srv.songs {
		if id.Protocol == p && id.Key == k {
			delete(srv.songs, id)
		}
	}
	srv.lock.Unlock()
	srv.Save()
	srv.broadcast(waitTracks)
	srv.broadcast(waitProtocols)
	return nil, nil
}

func (srv *Server) playlistChange(p Playlist, form url.Values, isq bool) (Playlist, error) {
	m := make([]*SongID, len(p))
	for i, v := range p {
		v := v
		m[i] = &v
	}
	for _, c := range form["c"] {
		sp := strings.SplitN(c, "-", 2)
		switch sp[0] {
		case "clear":
			for i := range m {
				m[i] = nil
			}
			if isq {
				go func() {
					srv.ch <- cmdStop
					srv.PlaylistIndex = 0
				}()
			}
		case "rem":
			i, err := strconv.Atoi(sp[1])
			if err != nil {
				return nil, err
			}
			if len(m) <= i {
				return nil, fmt.Errorf("unknown index: %v", i)
			}
			m[i] = nil
		case "add":
			id, err := ParseSongID(sp[1])
			if err != nil {
				return nil, err
			}
			m = append(m, &id)
		default:
			return nil, fmt.Errorf("unknown command: %v", sp[0])
		}
	}
	var pl Playlist
	for _, id := range m {
		if id != nil {
			pl = append(pl, *id)
		}
	}
	return pl, nil
}

func (srv *Server) QueueChange(form url.Values, ps httprouter.Params) (interface{}, error) {
	srv.lock.Lock()
	n, err := srv.playlistChange(srv.Queue, form, true)
	if err != nil {
		return nil, err
	} else {
		srv.Queue = n
		srv.Save()
	}
	srv.lock.Unlock()
	srv.broadcast(waitPlaylist)
	return nil, nil
}

func (srv *Server) PlaylistChange(form url.Values, ps httprouter.Params) (interface{}, error) {
	name := ps.ByName("playlist")
	p := srv.Playlists[name]
	srv.lock.Lock()
	n, err := srv.playlistChange(p, form, false)
	if err != nil {
		return nil, err
	} else {
		if len(n) == 0 {
			delete(srv.Playlists, name)
		} else {
			srv.Playlists[name] = n
		}
		srv.Save()
	}
	srv.lock.Unlock()
	srv.broadcast(waitPlaylist)
	return nil, nil
}

type listItem struct {
	ID   SongID
	Info *codec.SongInfo
}

func (srv *Server) list() []listItem {
	srv.lock.Lock()
	defer srv.lock.Unlock()
	songs := make([]listItem, len(srv.songs))
	i := 0
	for id, info := range srv.songs {
		songs[i] = listItem{
			ID:   id,
			Info: info,
		}
		i++
	}
	return songs
}

func (srv *Server) status() *Status {
	return &Status{
		State:    srv.state,
		Song:     srv.songID,
		SongInfo: srv.info,
		Elapsed:  srv.elapsed,
		Time:     srv.info.Time,
		Random:   srv.Random,
		Repeat:   srv.Repeat,
	}
}

type Status struct {
	// Playback state
	State State
	// Song ID.
	Song SongID
	SongInfo codec.SongInfo
	// Elapsed time of current song.
	Elapsed time.Duration
	// Duration of current song.
	Time   time.Duration
	Random bool
	Repeat bool
}

func serveError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
