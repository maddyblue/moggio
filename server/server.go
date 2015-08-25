// Package server implements the mog protocol.
package server

import (
	"bytes"
	"compress/gzip"
	crand "crypto/rand"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/big"
	"math/rand"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/protocol"
	"github.com/pkg/browser"
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
	if !devMode {
		host := addr
		if strings.HasPrefix(host, ":") {
			host = "localhost" + host
		}
		err := browser.OpenURL("http://" + host + "/")
		if err != nil {
			log.Println(err)
		}
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

type SongID codec.ID

func (s SongID) MarshalJSON() ([]byte, error) {
	sp := strings.SplitN(string(s), codec.IdSep, 3)
	if len(sp) != 3 {
		return json.Marshal("")
	}
	return json.Marshal(struct {
		Protocol string
		Key      string
		ID       string
		UID      string
	}{
		sp[0],
		sp[1],
		sp[2],
		string(s),
	})
}

func (s SongID) Protocol() string {
	return codec.ID(s).Top()
}

func (s SongID) Key() string {
	_, c := codec.ID(s).Pop()
	return c.Top()
}

func (s SongID) ID() codec.ID {
	_, c := codec.ID(s).Pop()
	_, c = c.Pop()
	return c
}

type Server struct {
	Queue     Playlist
	Playlists map[string]Playlist

	Repeat      bool
	Random      bool
	Protocols   map[string]map[string]protocol.Instance
	MinDuration time.Duration

	// Current song data.
	PlaylistIndex int
	songID        SongID
	song          codec.Song
	info          codec.SongInfo
	elapsed       time.Duration

	ch          chan interface{}
	audioch     chan interface{}
	state       State
	songs       map[SongID]*codec.SongInfo
	db          *bolt.DB
	savePending bool
}

func (srv *Server) removeDeleted(p Playlist) Playlist {
	var r Playlist
	for _, id := range p {
		if srv.songs[id] == nil {
			continue
		}
		r = append(r, id)
	}
	return r
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
		ch:          make(chan interface{}),
		audioch:     make(chan interface{}),
		songs:       make(map[SongID]*codec.SongInfo),
		Protocols:   make(map[string]map[string]protocol.Instance),
		Playlists:   make(map[string]Playlist),
		MinDuration: time.Second * 30,
	}
	for name := range protocol.Get() {
		srv.Protocols[name] = make(map[string]protocol.Instance)
	}
	db, err := bolt.Open(stateFile, 0600, nil)
	if err != nil {
		return nil, err
	}
	srv.db = db
	if err := srv.restore(); err != nil {
		log.Println(err)
	}
	log.Println("started from", stateFile)
	go srv.commands()
	go srv.audio()
	return &srv, nil
}

const (
	dbBucket = "bucket"
	dbServer = "server"
)

func (srv *Server) restore() error {
	decode := func(name string, dst interface{}) error {
		var data []byte
		err := srv.db.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(dbBucket))
			if b == nil {
				return fmt.Errorf("unknown bucket: %v", dbBucket)
			}
			data = b.Get([]byte(name))
			return nil
		})
		if err != nil {
			return err
		}
		gr, err := gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return err
		}
		defer gr.Close()
		return gob.NewDecoder(gr).Decode(dst)
	}
	if err := decode(dbServer, srv); err != nil {
		return err
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
	return nil
}

func (srv *Server) save() error {
	defer func() {
		srv.savePending = false
	}()
	store := map[string]interface{}{
		dbServer: srv,
	}
	tostore := make(map[string][]byte)
	for name, data := range store {
		f := new(bytes.Buffer)
		gz := gzip.NewWriter(f)
		enc := gob.NewEncoder(gz)
		if err := enc.Encode(data); err != nil {
			return err
		}
		if err := gz.Flush(); err != nil {
			return err
		}
		if err := gz.Close(); err != nil {
			return err
		}
		tostore[name] = f.Bytes()
	}
	err := srv.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte(dbBucket))
		if err != nil {
			return err
		}
		for name, data := range tostore {
			if err := b.Put([]byte(name), data); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	log.Println("save to db complete")
	return nil
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
	for k, v := range songs {
		if v.Time > 0 && v.Time < srv.MinDuration {
			delete(songs, k)
		}
	}
	srv.ch <- cmdRefresh{
		protocol: protocol,
		key:      key,
		songs:    songs,
	}
	return err
}

type PlaylistChange [][]string

func (srv *Server) playlistChange(p Playlist, plc PlaylistChange, isq bool) (pl Playlist, cleared bool, err error) {
	m := make([]SongID, len(p))
	copy(m, p)
	for _, c := range plc {
		cmd := c[0]
		var arg string
		if len(c) > 1 {
			arg = c[1]
		}
		switch cmd {
		case "clear":
			cleared = true
			for i := range m {
				m[i] = ""
			}
		case "rem":
			i, err := strconv.Atoi(arg)
			if err != nil {
				return nil, false, err
			}
			if len(m) <= i {
				return nil, false, fmt.Errorf("unknown index: %v", i)
			}
			m[i] = ""
		case "add":
			m = append(m, SongID(arg))
		default:
			return nil, false, fmt.Errorf("unknown command: %v", cmd)
		}
	}
	for _, id := range m {
		if id != "" {
			pl = append(pl, id)
		}
	}
	return
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
