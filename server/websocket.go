package server

import (
	"fmt"

	"github.com/mjibson/mog/protocol"
	"golang.org/x/net/websocket"
)

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
	waitError              = "error"
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
