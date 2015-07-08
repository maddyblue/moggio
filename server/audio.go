package server

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/mjibson/mog/_third_party/golang.org/x/net/websocket"
	"github.com/mjibson/mog/_third_party/golang.org/x/oauth2"
	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/output"
	"github.com/mjibson/mog/protocol"
)

func (srv *Server) audio() {
	var o output.Output
	var t chan interface{}
	var dur time.Duration
	srv.state = stateStop
	var next, stop, tick, play, pause, prev func()
	var timer <-chan time.Time
	waiters := make(map[*websocket.Conn]chan struct{})
	var seek *Seek
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
		wd, err := srv.makeWaitData(wt)
		if err != nil {
			log.Println(err)
			return
		}
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
	var forceNext = false
	stop = func() {
		log.Println("stop")
		srv.state = stateStop
		t = nil
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
			srv.info = *srv.songs[sid]
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
				broadcastErr(err)
				sendNext()
				return
			}
			o, err = output.Get(sr, ch)
			if err != nil {
				broadcastErr(fmt.Errorf("mog: could not open audio (%v, %v): %v", sr, ch, err))
				sendNext()
				return
			}
			srv.elapsed = 0
			dur = time.Second / (time.Duration(sr * ch))
			seek = NewSeek(srv.info.Time > 0, dur, srv.song.Play)
			log.Println("playing", srv.info.Title, sr, ch, dur, time.Duration(expected)*dur)
			t = make(chan interface{})
			close(t)
			srv.state = statePlay
		}
		next, err := seek.Read(expected)
		if err == nil {
			srv.elapsed = seek.Pos()
			if len(next) > 0 {
				o.Push(next)
			}
			select {
			case <-timer:
				// Check for updated song info.
				if info, err := inst.Info(sid.ID()); err != nil {
					broadcastErr(err)
				} else if srv.info != *info {
					srv.info = *info
					broadcast(waitStatus)
				}
				timer = nil
			default:
			}
			if timer == nil {
				timer = time.After(time.Second)
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
			if id.Protocol() == c.protocol && id.Key() == c.key {
				delete(srv.songs, id)
			}
		}
		for id, s := range c.songs {
			srv.songs[SongID(codec.NewID(c.protocol, c.key, string(id)))] = s
		}
		broadcast(waitTracks)
		broadcast(waitProtocols)
	}
	protocolRemove := func(c cmdProtocolRemove) {
		delete(c.prots, c.key)
		for id := range srv.songs {
			if id.Protocol() == c.protocol && id.Key() == c.key {
				delete(srv.songs, id)
			}
		}
		broadcast(waitTracks)
		broadcast(waitProtocols)
	}
	queueChange := func(c cmdQueueChange) {
		n, clear, err := srv.playlistChange(srv.Queue, PlaylistChange(c), true)
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
		n, _, err := srv.playlistChange(p, c.plc, false)
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
		go srv.protocolRefresh(c.name, instance.Key(), false)
		c.done <- nil
	}
	doSeek := func(c cmdSeek) {
		if seek == nil {
			return
		}
		err := seek.Seek(time.Duration(c))
		if err != nil {
			broadcastErr(err)
		}
	}
	setMinDuration := func(c cmdMinDuration) {
		srv.MinDuration = time.Duration(c)
	}
	ch := make(chan interface{})
	go func() {
		for c := range srv.ch {
			timer := time.AfterFunc(time.Second*10, func() {
				panic("delay timer expired")
			})
			ch <- c
			timer.Stop()
		}
	}()
	for {
		select {
		case <-t:
			tick()
		case c := <-ch:
			save := true
			log.Printf("%T\n", c)
			switch c := c.(type) {
			case controlCmd:
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

type cmdSeek time.Duration

type cmdPlayIdx int

type cmdRefresh struct {
	protocol, key string
	songs         protocol.SongList
}

type cmdProtocolRemove struct {
	protocol, key string
	prots         map[string]protocol.Instance
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
