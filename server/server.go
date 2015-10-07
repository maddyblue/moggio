// Package server implements the mog protocol.
package server

import (
	"bytes"
	"compress/gzip"
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
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/boltdb/bolt"
	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/models"
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
	println(string(b))
}

func ListenAndServe(stateFile, addr, central string, devMode bool) error {
	server, err := New(stateFile, central)
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

func (s SongID) Triple() (protocol, key string, id codec.ID) {
	protocol, id = codec.ID(s).Pop()
	key, id = id.Pop()
	return
}

type Server struct {
	Queue     Playlist
	Playlists map[string]Playlist

	Username string
	Token    string

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

	centralURL  string
	inprogress  map[codec.ID]bool
	ch          chan interface{}
	audioch     chan interface{}
	state       State
	db          *bolt.DB
	savePending bool
}

func (srv *Server) removeDeleted(p Playlist) Playlist {
	var r Playlist
	for _, id := range p {
		if !srv.hasSong(id) {
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
		info, _ := srv.getSong(id)
		r[idx] = listItem{
			ID:   id,
			Info: info,
		}
	}
	return r
}

var dir = filepath.Join("server")

func New(stateFile, central string) (*Server, error) {
	srv := Server{
		ch:          make(chan interface{}),
		audioch:     make(chan interface{}),
		Protocols:   protocol.Map(),
		Playlists:   make(map[string]Playlist),
		MinDuration: time.Second * 30,
		centralURL:  central,
		inprogress:  make(map[codec.ID]bool),
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

func (srv *Server) protocolRefresh(protocol, key string, list, doDelete bool) error {
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
	if doDelete {
		srv.ch <- cmdRemoveDeleted{}
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
	Time       time.Duration
	Random     bool
	Repeat     bool
	Username   string
	Hostname   string
	CentralURL string
}

func (srv *Server) request(path string, body interface{}) (io.ReadCloser, error) {
	// TODO: srv.Token is subject to a race condition because this function is
	// called in go routines in the control loop, and srv.Token is set in the
	// main go routine.
	tv := url.Values{"token": []string{srv.Token}}.Encode()
	var br io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		br = bytes.NewReader(b)
	}
	r, err := http.Post(srv.centralURL+path+"?"+tv, "application/json", br)
	if err != nil {
		return nil, err
	}
	if r.StatusCode != 200 {
		b, err := ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("%s: %v: %s", path, r.Status, b)
	}
	return r.Body, nil
}

// name and key may be empty to match all
func (srv *Server) putSource(protocol, key string) error {
	var ss []*models.Source
	for prot, m := range srv.Protocols {
		if prot != protocol && protocol != "" {
			continue
		}
		for name, p := range m {
			if name != key && key != "" {
				continue
			}
			buf := new(bytes.Buffer)
			gw := gzip.NewWriter(buf)
			if err := gob.NewEncoder(gw).Encode(p); err != nil {
				return err
			}
			if err := gw.Close(); err != nil {
				return err
			}
			ss = append(ss, &models.Source{
				Protocol: prot,
				Name:     name,
				Blob:     buf.Bytes(),
			})
		}
	}
	r, err := srv.request("/api/source/set", &ss)
	if err != nil {
		return fmt.Errorf("could not set sources: %v", err)
	}
	r.Close()
	return nil
}

func (srv *Server) getSong(id SongID) (*codec.SongInfo, error) {
	name, key, cid := id.Triple()
	p, ok := srv.Protocols[name]
	if !ok {
		return nil, fmt.Errorf("unknown protocol: %s", name)
	}
	inst, ok := p[key]
	if !ok {
		return nil, fmt.Errorf("unknown instance: %s", key)
	}
	return inst.Info(cid)
}

func (srv *Server) hasSong(id SongID) bool {
	info, _ := srv.getSong(id)
	return info != nil
}
