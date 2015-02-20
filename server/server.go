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
	"net/url"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/mjibson/mog/_third_party/github.com/boltdb/bolt"
	"github.com/mjibson/mog/_third_party/github.com/pkg/browser"
	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/protocol"
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
	db          *bolt.DB
	savePending bool
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
	db, err := bolt.Open(stateFile, 0600, nil)
	if err != nil {
		return nil, err
	}
	srv.db = db
	if err := srv.restore(); err != nil {
		log.Println(err)
	}
	log.Println("started from", stateFile)
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
	srv.ch <- cmdRefresh{
		protocol: protocol,
		key:      key,
		songs:    songs,
	}
	return err
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
