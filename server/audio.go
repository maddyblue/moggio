package server

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/hajimehoshi/oto"
)

func (srv *Server) audio() {
	var player *oto.Player
	var t chan interface{}
	var seek *Seek
	var dur time.Duration
	var err error
	send := func(v interface{}) {
		go func() {
			srv.ch <- v
		}()
	}
	setTime := func(force bool) {
		send(cmdSetTime{
			duration: seek.Pos(),
			force:    force,
		})
	}
	const expected = 4096
	tick := func() {
		if seek == nil {
			return
		}
		next, err := seek.Read(expected)
		if len(next) > 0 {
			// TODO: check for error?
			player.Write(next)
			setTime(false)
		}
		if err != nil {
			seek = nil
		}
		if err == io.ErrUnexpectedEOF {
			send(cmdRestartSong)
		} else if err != nil {
			send(cmdNext)
		}
	}
	doSeek := func(c cmdSeek) {
		if seek == nil {
			return
		}
		err := seek.Seek(time.Duration(c))
		if err != nil {
			send(cmdError(err))
			return
		}
		setTime(true)
	}
	setParams := func(c audioSetParams) {
		player, err = oto.NewPlayer(c.sr, c.ch, 2, expected)
		if err != nil {
			c.err <- fmt.Errorf("moggio: could not open audio (%v, %v): %v", c.sr, c.ch, err)
			return
		}
		dur = time.Second / (time.Duration(c.sr * c.ch))
		seek = NewSeek(c.dur > 0, dur, c.play)
		t = make(chan interface{})
		close(t)
		c.err <- nil
	}
	for {
		select {
		case <-t:
			tick()
		case c := <-srv.audioch:
			log.Printf("%T\n", c)
			switch c := c.(type) {
			case audioStop:
				t = nil
			case audioPlay:
				t = make(chan interface{})
				close(t)
			case audioSetParams:
				setParams(c)
			case cmdSeek:
				doSeek(c)
			default:
				panic("unknown type")
			}
		}
	}
}

type audioSetParams struct {
	sr   int
	ch   int
	dur  time.Duration
	play io.Reader
	err  chan error
}

type audioStop struct{}

type audioPlay struct{}
