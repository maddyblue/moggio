// Package server implements the mog protocol.
package server

import (
	crand "crypto/rand"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/big"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/output"
	"github.com/mjibson/mog/protocol"
	"golang.org/x/net/websocket"
	"golang.org/x/oauth2"
)

func init() {
	i, err := crand.Int(crand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		return
	}
	rand.Seed(i.Int64())
}

func printErr(e error) {
	log.Println(e)
	b := make([]byte, 4096)
	runtime.Stack(b, false)
	fmt.Println(string(b))
}

func ListenAndServe(stateFile, addr string, devMode bool) error {
	server, err := New(stateFile)
	if err != nil {
		return err
	}
	return server.ListenAndServe(addr, devMode)
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

	ch          chan interface{}
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

// makeWaitData should only be called by the audio() function.
func (srv *Server) makeWaitData(wt waitType) (*waitData, error) {
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
		data = &Status{
			State:    srv.state,
			Song:     srv.songID,
			SongInfo: srv.info,
			Elapsed:  srv.elapsed,
			Time:     srv.info.Time,
			Random:   srv.Random,
			Repeat:   srv.Repeat,
		}
	case waitTracks:
		songs := make([]listItem, len(srv.songs))
		i := 0
		for id, info := range srv.songs {
			songs[i] = listItem{
				ID:   id,
				Info: info,
			}
			i++
		}
		data = struct {
			Tracks []listItem
		}{
			Tracks: songs,
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
		return nil, fmt.Errorf("bad wait type: %s", wt)
	}
	return &waitData{
		Type: wt,
		Data: data,
	}, nil
}

type PlaylistInfo []listItem

func (srv *Server) playlistInfo(p Playlist) PlaylistInfo {
	r := make(PlaylistInfo, len(p))
	for idx, id := range p {
		r[idx] = listItem{
			ID:   id,
			Info: srv.songs[id],
		}
	}
	return r
}

var dir = filepath.Join("server")

func New(stateFile string) (*Server, error) {
	srv := Server{
		ch:        make(chan interface{}),
		songs:     make(map[SongID]*codec.SongInfo),
		Protocols: make(map[string]map[string]protocol.Instance),
		Playlists: make(map[string]Playlist),
	}
	for name := range protocol.Get() {
		srv.Protocols[name] = make(map[string]protocol.Instance)
	}
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
	router.GET("/api/data/:type", JSON(srv.Data))
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
func (srv *Server) ListenAndServe(addr string, devMode bool) error {
	mux := srv.GetMux(devMode)
	log.Println("mog: listening on", addr)
	return http.ListenAndServe(addr, mux)
}

type cmdNewWS struct {
	ws   *websocket.Conn
	done chan struct{}
}

type cmdDeleteWS *websocket.Conn

func (srv *Server) WebSocket(ws *websocket.Conn) {
	c := make(chan struct{})
	srv.ch <- cmdNewWS{
		ws:   ws,
		done: c,
	}
	for range c {
	}
}

func (srv *Server) error(err error) {
	// TODO: broadcast err
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
	waiters := make(map[*websocket.Conn]chan struct{})
	broadcast := func(wt waitType) {
		wd, err := srv.makeWaitData(wt)
		if err != nil {
			log.Println(err)
			return
		}
		for ws := range waiters {
			go func(ws *websocket.Conn) {
				if err := websocket.JSON.Send(ws, wd); err != nil {
					srv.ch <- cmdDeleteWS(ws)
				}
			}(ws)
		}
	}
	newWS := func(c cmdNewWS) {
		ws := (*websocket.Conn)(c.ws)
		waiters[ws] = c.done
		inits := []waitType{
			waitPlaylist,
			waitProtocols,
			waitStatus,
			waitTracks,
		}
		for _, wt := range inits {
			data, err := srv.makeWaitData(wt)
			if err != nil {
				return
			}
			go func() {
				if err := websocket.JSON.Send(ws, data); err != nil {
					srv.ch <- cmdDeleteWS(ws)
					return
				}
			}()
		}
	}
	deleteWS := func(c cmdDeleteWS) {
		ws := (*websocket.Conn)(c)
		ch := waiters[ws]
		if ch == nil {
			return
		}
		close(ch)
		delete(waiters, ws)
	}
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
				broadcast(waitStatus)
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
	playIdx := func(c cmdPlayIdx) {
		stop()
		srv.PlaylistIndex = int(c)
		play()
	}
	refresh := func(c cmdRefresh) {
		for id := range srv.songs {
			if id.Protocol == c.protocol {
				delete(srv.songs, id)
			}
		}
		for id, s := range c.songs {
			srv.songs[SongID{
				Protocol: c.protocol,
				Key:      c.key,
				ID:       id,
			}] = s
		}
		broadcast(waitTracks)
		broadcast(waitProtocols)
	}
	protocolRemove := func(c cmdProtocolRemove) {
		delete(c.prots, c.key)
		for id := range srv.songs {
			if id.Protocol == c.protocol && id.Key == c.key {
				delete(srv.songs, id)
			}
		}
		broadcast(waitTracks)
		broadcast(waitProtocols)
	}
	queueChange := func(c cmdQueueChange) {
		n, clear, err := srv.playlistChange(srv.Queue, url.Values(c), true)
		if err != nil {
			srv.error(err)
			return
		}
		srv.Queue = n
		if clear || len(n) == 0 {
			stop()
			srv.PlaylistIndex = 0
		}
		broadcast(waitPlaylist)
	}
	playlistChange := func(c cmdPlaylistChange) {
		p := srv.Playlists[c.name]
		n, _, err := srv.playlistChange(p, c.form, false)
		if err != nil {
			srv.error(err)
			return
		}
		if len(n) == 0 {
			delete(srv.Playlists, c.name)
		} else {
			srv.Playlists[c.name] = n
		}
		broadcast(waitPlaylist)
	}
	queueSave := func() {
		if srv.stateFile == "" {
			return
		}
		if srv.savePending {
			return
		}
		srv.savePending = true
		time.AfterFunc(time.Second, func() {
			srv.ch <- cmdDoSave{}
		})
	}
	doSave := func() {
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
	addOAuth := func(c cmdAddOAuth) {
		prot, err := protocol.ByName(c.name)
		if err != nil {
			c.done <- err
			return
		}
		prots, ok := srv.Protocols[c.name]
		if !ok || prot.OAuth == nil {
			c.done <- fmt.Errorf("bad protocol")
			return
		}
		// TODO: decouple this from the audio thread
		t, err := prot.OAuth.Exchange(oauth2.NoContext, c.r.FormValue("code"))
		if err != nil {
			c.done <- err
			return
		}
		// "Bearer" was added for dropbox. It happens to work also with Google Music's
		// OAuth. This may need to be changed to be protocol-specific in the future.
		t.TokenType = "Bearer"
		instance, err := prot.NewInstance(nil, t)
		if err != nil {
			c.done <- err
			return
		}
		prots[t.AccessToken] = instance
		queueSave()
		go srv.protocolRefresh(c.name, instance.Key(), false)
		c.done <- nil
	}
	for {
		select {
		case <-t:
			tick()
		case c := <-srv.ch:
			save := true
			log.Printf("%T\n", c)
			switch c := c.(type) {
			case controlCmd:
				switch c {
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
				case cmdRandom:
					srv.Random = !srv.Random
				case cmdRepeat:
					srv.Repeat = !srv.Repeat
				default:
					panic(c)
				}
			case cmdPlayIdx:
				playIdx(c)
			case cmdRefresh:
				refresh(c)
			case cmdProtocolRemove:
				protocolRemove(c)
			case cmdQueueChange:
				queueChange(c)
			case cmdPlaylistChange:
				playlistChange(c)
			case cmdNewWS:
				newWS(c)
			case cmdDeleteWS:
				deleteWS(c)
			case cmdQueueSave:
				queueSave()
			case cmdDoSave:
				save = false
				doSave()
			case cmdAddOAuth:
				addOAuth(c)
			default:
				panic(c)
			}
			broadcast(waitStatus)
			if save {
				queueSave()
			}
		}
	}
}

type controlCmd int

const (
	cmdUnknown controlCmd = iota
	cmdNext
	cmdPause
	cmdPlay
	cmdPrev
	cmdRandom
	cmdRepeat
	cmdStop
)

type cmdPlayIdx int

type cmdRefresh struct {
	protocol, key string
	songs         protocol.SongList
}

type cmdProtocolRemove struct {
	protocol, key string
	prots         map[string]protocol.Instance
}

type cmdQueueChange url.Values

type cmdPlaylistChange struct {
	form url.Values
	name string
}

type cmdQueueSave struct{}

type cmdDoSave struct{}

type cmdAddOAuth struct {
	name string
	r    *http.Request
	done chan error
}

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

func (srv *Server) Data(form url.Values, ps httprouter.Params) (interface{}, error) {
	return srv.makeWaitData(waitType(ps.ByName("type")))
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
		srv.ch <- cmdPlayIdx(i)
	case "random":
		srv.ch <- cmdRandom
	case "repeat":
		srv.ch <- cmdRepeat
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
	srv.ch <- cmdRefresh{
		protocol: protocol,
		key:      key,
		songs:    songs,
	}
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
	prots, ok := srv.Protocols[p]
	if !ok {
		return nil, fmt.Errorf("unknown protocol: %v", p)
	}
	srv.ch <- cmdProtocolRemove{
		protocol: p,
		key:      k,
		prots:    prots,
	}
	return nil, nil
}

func (srv *Server) playlistChange(p Playlist, form url.Values, isq bool) (pl Playlist, cleared bool, err error) {
	m := make([]*SongID, len(p))
	for i, v := range p {
		v := v
		m[i] = &v
	}
	for _, c := range form["c"] {
		sp := strings.SplitN(c, "-", 2)
		switch sp[0] {
		case "clear":
			cleared = true
			for i := range m {
				m[i] = nil
			}
		case "rem":
			i, err := strconv.Atoi(sp[1])
			if err != nil {
				return nil, false, err
			}
			if len(m) <= i {
				return nil, false, fmt.Errorf("unknown index: %v", i)
			}
			m[i] = nil
		case "add":
			id, err := ParseSongID(sp[1])
			if err != nil {
				return nil, false, err
			}
			m = append(m, &id)
		default:
			return nil, false, fmt.Errorf("unknown command: %v", sp[0])
		}
	}
	for _, id := range m {
		if id != nil {
			pl = append(pl, *id)
		}
	}
	return
}

func (srv *Server) QueueChange(form url.Values, ps httprouter.Params) (interface{}, error) {
	srv.ch <- cmdQueueChange(form)
	return nil, nil
}

func (srv *Server) PlaylistChange(form url.Values, ps httprouter.Params) (interface{}, error) {
	srv.ch <- cmdPlaylistChange{
		form: form,
		name: ps.ByName("playlist"),
	}
	return nil, nil
}

type listItem struct {
	ID   SongID
	Info *codec.SongInfo
}

type Status struct {
	// Playback state
	State State
	// Song ID.
	Song     SongID
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
