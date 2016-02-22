package server

import (
	"bytes"
	"compress/gzip"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/bradfitz/slice"
	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/models"
	"github.com/mjibson/mog/protocol"
	"golang.org/x/net/websocket"
	"golang.org/x/oauth2"
)

func (srv *Server) commands() {
	srv.state = stateStop
	var next, stop, tick, play, pause, prev func()
	var timer <-chan time.Time
	waiters := make(map[*websocket.Conn]chan struct{})
	broadcastData := func(wd *waitData) {
		for ws := range waiters {
			go func(ws *websocket.Conn) {
				if err := websocket.JSON.Send(ws, wd); err != nil {
					srv.ch <- cmdDeleteWS(ws)
				}
			}(ws)
		}
	}
	broadcast := func(wt waitType) {
		wd := srv.makeWaitData(wt)
		broadcastData(wd)
	}
	broadcastErr := func(err error) {
		printErr(err)
		v := struct {
			Time  time.Time
			Error string
		}{
			time.Now().UTC(),
			err.Error(),
		}
		broadcastData(&waitData{
			Type: waitError,
			Data: v,
		})
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
			data := srv.makeWaitData(wt)
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
			srv.audioch <- audioPlay{}
			srv.state = statePlay
		case statePlay:
			log.Println("pause: pause")
			srv.audioch <- audioStop{}
			srv.state = statePause
		}
	}
	next = func() {
		log.Println("next")
		stop()
		play()
	}
	var forceNext = false
	stop = func() {
		log.Println("stop")
		srv.state = stateStop
		srv.audioch <- audioStop{}
		if srv.song != nil || forceNext {
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
		forceNext = false
		srv.song = nil
		srv.elapsed = 0
	}
	var inst protocol.Instance
	var sid SongID
	sendNext := func() {
		go func() {
			srv.ch <- cmdNext
		}()
	}
	nextOpen := time.After(0)
	tick = func() {
		const expected = 4096
		if false && srv.elapsed > srv.info.Time {
			log.Println("elapsed time completed", srv.elapsed, srv.info.Time)
			stop()
		}
		if srv.song == nil {
			<-nextOpen
			nextOpen = time.After(time.Second / 2)
			defer broadcast(waitStatus)
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
			sid = srv.songID
			if info, err := srv.getSong(sid); err != nil {
				broadcastErr(err)
				forceNext = true
				sendNext()
				return
			} else {
				srv.info = *info
			}
			inst = srv.Protocols[sid.Protocol()][sid.Key()]
			song, err := inst.GetSong(sid.ID())
			if err != nil {
				forceNext = true
				broadcastErr(err)
				sendNext()
				return
			}
			srv.song = song
			sr, ch, err := srv.song.Init()
			if err != nil {
				srv.song.Close()
				srv.song = nil
				broadcastErr(err)
				sendNext()
				return
			}
			params := audioSetParams{
				sr:   sr,
				ch:   ch,
				dur:  srv.info.Time,
				play: srv.song.Play,
				err:  make(chan error),
			}
			srv.audioch <- params
			if err := <-params.err; err != nil {
				broadcastErr(err)
				sendNext()
				return
			}
			srv.elapsed = 0
			log.Println("playing", srv.info.Title, sr, ch)
			srv.state = statePlay
		}
	}
	infoTimer := func() {
		timer = time.After(time.Second)
		if inst == nil {
			return
		}
		// Check for updated song info.
		if info, err := inst.Info(sid.ID()); err != nil {
			broadcastErr(err)
		} else if srv.info != *info {
			srv.info = *info
			broadcast(waitStatus)
		}
	}
	restart := func() {
		log.Println("attempting to restart song")
		n := srv.PlaylistIndex
		stop()
		srv.PlaylistIndex = n
		play()
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
	playTrack := func(c cmdPlayTrack) {
		t := SongID(c)
		info, err := srv.getSong(t)
		if err != nil {
			broadcastErr(err)
			return
		}
		album := info.Album
		p, err := srv.getInstance(t.Protocol(), t.Key())
		if err != nil {
			broadcastErr(err)
			return
		}
		list, err := p.List()
		if err != nil {
			broadcastErr(err)
			return
		}
		top := codec.NewID(t.Protocol(), t.Key())
		var ids []codec.ID
		for id, si := range list {
			if si.Album == album {
				ids = append(ids, id)
			}
		}
		slice.Sort(ids, func(i, j int) bool {
			a := list[ids[i]]
			b := list[ids[j]]
			return a.Track < b.Track
		})
		plc := PlaylistChange{[]string{"clear"}}
		for _, v := range ids {
			plc = append(plc, []string{"add", string(top.Push(string(v)))})
		}
		n, _, err := srv.playlistChange(srv.Queue, plc)
		if err != nil {
			broadcastErr(err)
			return
		}
		stop()
		srv.Queue = n
		srv.PlaylistIndex = 0
		for i, s := range srv.Queue {
			if s == t {
				srv.PlaylistIndex = i
			}
		}
		play()
		broadcast(waitPlaylist)
	}
	removeDeleted := func() {
		for n, p := range srv.Playlists {
			p = srv.removeDeleted(p)
			if len(p) == 0 {
				delete(srv.Playlists, n)
			} else {
				srv.Playlists[n] = p
			}
		}
		srv.Queue = srv.removeDeleted(srv.Queue)
		if info, _ := srv.getSong(srv.songID); info == nil {
			playing := srv.state == statePlay
			stop()
			if playing {
				srv.PlaylistIndex = 0
				play()
			}
		}
		broadcast(waitPlaylist)
	}
	protocolRemove := func(c cmdProtocolRemove) {
		prots, ok := srv.Protocols[c.protocol]
		if !ok {
			return
		}
		delete(prots, c.key)
		if srv.Token != "" {
			d := models.Delete{
				Protocol: c.protocol,
				Name:     c.key,
			}
			go func() {
				r, err := srv.request("/api/source/delete", &d)
				if err != nil {
					srv.ch <- cmdError(err)
					return
				}
				r.Close()
			}()
		}
		broadcast(waitTracks)
		broadcast(waitProtocols)
	}
	removeInProgress := func(c cmdRemoveInProgress) {
		delete(srv.inprogress, codec.ID(c))
		broadcast(waitProtocols)
	}
	protocolAdd := func(c cmdProtocolAdd) {
		name, key := c.Name, c.Instance.Key()
		id := codec.NewID(name, key)
		if srv.inprogress[id] {
			broadcastErr(fmt.Errorf("already adding %s: %s", name, key))
			return
		}
		if _, err := srv.getInstance(name, key); err == nil {
			broadcastErr(fmt.Errorf("already have %s: %s", name, key))
			return
		}
		srv.inprogress[id] = true
		broadcast(waitProtocols)
		go func() {
			defer func() {
				srv.ch <- cmdRemoveInProgress(id)
			}()
			songs, err := c.Instance.Refresh()
			if err != nil {
				srv.ch <- cmdError(err)
				return
			}
			for k, v := range songs {
				if v.Time > 0 && v.Time < srv.MinDuration {
					delete(songs, k)
				}
			}
			srv.ch <- cmdProtocolAddInstance(c)
		}()
	}
	protocolAddInstance := func(c cmdProtocolAddInstance) {
		srv.Protocols[c.Name][c.Instance.Key()] = c.Instance
		if srv.Token != "" {
			srv.ch <- cmdPutSource{
				protocol: c.Name,
				key:      c.Instance.Key(),
			}
		}
		broadcast(waitTracks)
		broadcast(waitProtocols)
	}
	queueChange := func(c cmdQueueChange) {
		n, clear, err := srv.playlistChange(srv.Queue, PlaylistChange(c))
		if err != nil {
			broadcastErr(err)
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
		n, _, err := srv.playlistChange(p, c.plc)
		if err != nil {
			broadcastErr(err)
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
		if srv.savePending {
			return
		}
		srv.savePending = true
		time.AfterFunc(time.Second, func() {
			srv.ch <- cmdDoSave{}
		})
	}
	doSave := func() {
		if err := srv.save(); err != nil {
			broadcastErr(err)
		}
	}
	addOAuth := func(c cmdAddOAuth) {
		go func() {
			prot, err := protocol.ByName(c.name)
			if err != nil {
				c.done <- err
				return
			}
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
			srv.ch <- cmdProtocolAdd{
				Name:     c.name,
				Instance: instance,
			}
			c.done <- nil
		}()
	}
	setMinDuration := func(c cmdMinDuration) {
		srv.MinDuration = time.Duration(c)
	}
	doSeek := func(c cmdSeek) {
		if time.Duration(c) > srv.info.Time {
			return
		}
		srv.audioch <- c
	}
	setUsername := func(c cmdSetUsername) {
		srv.Username = string(c)
	}
	makeSource := func(protocol, key string) ([]*models.Source, error) {
		// name and key may be empty to match all
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
					return nil, err
				}
				if err := gw.Close(); err != nil {
					return nil, err
				}
				ss = append(ss, &models.Source{
					Protocol: prot,
					Name:     name,
					Blob:     buf.Bytes(),
				})
			}
		}
		return ss, nil
	}
	sendSource := func(ss []*models.Source) error {
		r, err := srv.request("/api/source/set", &ss)
		if err != nil {
			return err
		}
		r.Close()
		return nil
	}
	putSource := func(c cmdPutSource) {
		ss, err := makeSource(c.protocol, c.key)
		if err != nil {
			broadcastErr(fmt.Errorf("could not set sources: %v", err))
			return
		}
		if err := sendSource(ss); err != nil {
			broadcastErr(fmt.Errorf("could not set sources: %v", err))
		}
	}
	tokenRegister := func(c cmdTokenRegister) {
		srv.Token = ""
		if c != "" {
			srv.Token = string(c)
			ss, err := makeSource("", "")
			if err != nil {
				broadcastErr(err)
				return
			}
			go func() {
				if err := sendSource(ss); err != nil {
					srv.ch <- cmdError(err)
					return
				}
				r, err := srv.request("/api/source/get", nil)
				if err != nil {
					srv.ch <- cmdError(fmt.Errorf("could not get sources: %v", err))
					return
				}
				defer r.Close()
				if err := json.NewDecoder(r).Decode(&ss); err != nil {
					srv.ch <- cmdError(err)
					return
				}
				srv.ch <- cmdSetSources(ss)
			}()
		}
		go func() {
			srv.ch <- cmdSetUsername("")
			if c == "" {
				return
			}
			r, err := srv.request("/api/username", nil)
			if err != nil {
				srv.ch <- cmdError(err)
				return
			}
			defer r.Close()
			var u cmdSetUsername
			if err := json.NewDecoder(r).Decode(&u); err != nil {
				srv.ch <- cmdError(err)
				return
			}
			srv.ch <- u
		}()
	}
	setSources := func(c cmdSetSources) {
		ps := protocol.Map()
		for _, s := range c {
			proto, err := protocol.ByName(s.Protocol)
			if err != nil {
				broadcastErr(err)
				return
			}
			if _, ok := ps[s.Protocol]; !ok {
				ps[s.Protocol] = make(map[string]protocol.Instance)
			}
			r, err := gzip.NewReader(bytes.NewReader(s.Blob))
			if err != nil {
				broadcastErr(err)
				return
			}
			defer r.Close()
			p, err := proto.Decode(r)
			if err != nil {
				broadcastErr(err)
				return
			}
			ps[s.Protocol][s.Name] = p
		}
		srv.Protocols = ps
		go func() {
			// protocolRefresh uses srv.Protocols, so
			for i, s := range c {
				last := i == len(c)-1
				srv.ch <- cmdProtocolRefresh{
					protocol: s.Protocol,
					key:      s.Name,
					list:     true,
					doDelete: last,
					err:      make(chan error, 1),
				}
			}
		}()
	}
	sendWaitData := func(c cmdWaitData) {
		c.done <- srv.makeWaitData(c.wt)
	}
	protocolRefresh := func(c cmdProtocolRefresh) {
		id := codec.NewID(c.protocol, c.key)
		if srv.inprogress[id] {
			c.err <- nil
			return
		}
		inst, err := srv.getInstance(c.protocol, c.key)
		if err != nil {
			c.err <- err
			return
		}
		srv.inprogress[id] = true
		broadcast(waitProtocols)
		go func() {
			defer func() {
				srv.ch <- cmdRemoveInProgress(id)
			}()
			f := inst.Refresh
			if c.list {
				f = inst.List
			}
			songs, err := f()
			if err != nil {
				c.err <- err
				return
			}
			for k, v := range songs {
				if v.Time > 0 && v.Time < srv.MinDuration {
					delete(songs, k)
				}
			}
			if c.doDelete {
				srv.ch <- cmdRemoveDeleted{}
			}
			if srv.Token != "" {
				srv.ch <- cmdPutSource{
					protocol: c.protocol,
					key:      c.key,
				}
			}
			c.err <- nil
		}()
	}
	ch := make(chan interface{})
	go func() {
		for c := range srv.ch {
			go func(c interface{}) {
				timer := time.AfterFunc(time.Second*10, func() {
					log.Printf("%T: %#v\n", c, c)
					panic("delay timer expired")
				})
				ch <- c
				timer.Stop()
			}(c)
		}
	}()
	infoTimer()
	for {
		select {
		case <-timer:
			infoTimer()
		case c := <-ch:
			if c, ok := c.(cmdSetTime); ok {
				d := c.duration
				change := srv.elapsed - d
				if change < 0 {
					change = -change
				}
				srv.elapsed = d
				if c.force || change > time.Second {
					broadcast(waitStatus)
				}
				continue
			}
			save := true
			doDroadcast := false
			log.Printf("%T\n", c)
			switch c := c.(type) {
			case controlCmd:
				doDroadcast = true
				switch c {
				case cmdPlay:
					save = false
					play()
				case cmdStop:
					save = false
					stop()
				case cmdNext:
					next()
				case cmdPause:
					save = false
					pause()
				case cmdPrev:
					prev()
				case cmdRandom:
					srv.Random = !srv.Random
				case cmdRepeat:
					srv.Repeat = !srv.Repeat
				case cmdRestartSong:
					restart()
				default:
					panic(c)
				}
			case cmdPlayIdx:
				playIdx(c)
			case cmdPlayTrack:
				playTrack(c)
			case cmdProtocolRemove:
				protocolRemove(c)
			case cmdQueueChange:
				queueChange(c)
			case cmdPlaylistChange:
				playlistChange(c)
			case cmdNewWS:
				save = false
				newWS(c)
			case cmdDeleteWS:
				save = false
				deleteWS(c)
			case cmdDoSave:
				save = false
				doSave()
			case cmdAddOAuth:
				addOAuth(c)
			case cmdSeek:
				save = false
				doSeek(c)
			case cmdMinDuration:
				setMinDuration(c)
			case cmdTokenRegister:
				tokenRegister(c)
			case cmdSetUsername:
				setUsername(c)
			case cmdSetSources:
				setSources(c)
			case cmdProtocolAdd:
				protocolAdd(c)
			case cmdProtocolAddInstance:
				protocolAddInstance(c)
			case cmdRemoveDeleted:
				removeDeleted()
			case cmdRemoveInProgress:
				removeInProgress(c)
			case cmdError:
				broadcastErr(error(c))
				save = false
			case cmdWaitData:
				sendWaitData(c)
				save = false
			case cmdPutSource:
				putSource(c)
				save = false
			case cmdProtocolRefresh:
				protocolRefresh(c)
			default:
				panic(c)
			}
			if save {
				queueSave()
			}
			if save || doDroadcast {
				broadcast(waitStatus)
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
	cmdRestartSong
)

type cmdSeek time.Duration

type cmdPlayIdx int

type cmdRemoveDeleted struct{}

type cmdProtocolRemove struct {
	protocol, key string
}

type cmdQueueChange PlaylistChange

type cmdPlaylistChange struct {
	plc  PlaylistChange
	name string
}

type cmdDoSave struct{}

type cmdAddOAuth struct {
	name string
	r    *http.Request
	done chan error
}

type cmdMinDuration time.Duration

type cmdSetTime struct {
	duration time.Duration
	force    bool
}

type cmdError error

type cmdTokenRegister string

type cmdSetUsername string

type cmdSetSources []*models.Source

type cmdProtocolAdd struct {
	Name     string
	Instance protocol.Instance
}

type cmdProtocolAddInstance cmdProtocolAdd

type cmdRemoveInProgress codec.ID

type cmdWaitData struct {
	wt   waitType
	done chan<- *waitData
}

type cmdPutSource struct {
	protocol, key string
}

type cmdProtocolRefresh struct {
	protocol, key  string
	list, doDelete bool
	err            chan error
}

type cmdPlayTrack SongID
