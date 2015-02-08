package gmusic

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/mjibson/mog/_third_party/golang.org/x/oauth2"
	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/codec/mpa"
	"github.com/mjibson/mog/protocol"
	"github.com/mjibson/mog/protocol/gmusic/gmusic"
)

func init() {
	gob.Register(new(GMusic))
	protocol.Register("gmusic", []string{"username", "password"}, New)
}

func New(params []string, token *oauth2.Token) (protocol.Instance, error) {
	if len(params) != 2 {
		return nil, fmt.Errorf("gmusic: bad params")
	}
	g, err := gmusic.Login(params[0], params[1])
	if err != nil {
		return nil, err
	}
	return &GMusic{
		GMusic: g,
	}, nil
}

func (g *GMusic) Info(id string) (*codec.SongInfo, error) {
	s := g.Songs[id]
	if s == nil {
		return nil, fmt.Errorf("could not find %v", id)
	}
	return s, nil
}

type GMusic struct {
	GMusic *gmusic.GMusic
	Tracks map[string]*gmusic.Track
	Songs  protocol.SongList
}

func (g *GMusic) Key() string {
	return g.GMusic.DeviceID
}

func (g *GMusic) List(progress chan<- protocol.SongList) (protocol.SongList, error) {
	if len(g.Songs) == 0 {
		return g.Refresh(progress)
	}
	return g.Songs, nil
}

func (g *GMusic) GetSong(id string) (codec.Song, error) {
	f := g.Tracks[id]
	if f == nil {
		return nil, fmt.Errorf("missing %v", id)
	}
	return mpa.NewSong(func() (io.ReadCloser, int64, error) {
		log.Println("GMUSIC", id)
		r, err := g.GMusic.GetStream(id)
		if err != nil {
			return nil, 0, err
		}
		size, _ := strconv.ParseInt(f.EstimatedSize, 10, 64)
		return r.Body, size, nil
	})
}

func (g *GMusic) Refresh(progress chan<- protocol.SongList) (protocol.SongList, error) {
	tracks := make(map[string]*gmusic.Track)
	songs := make(protocol.SongList)
	log.Println("get gmusic tracks")
	trackList, err := g.GMusic.ListTracks()
	if err != nil {
		return nil, err
	}
	log.Println("got gmusic tracks", len(trackList))
	for _, t := range trackList {
		tracks[t.ID] = t
		duration, _ := strconv.Atoi(t.DurationMillis)
		songs[t.ID] = &codec.SongInfo{
			Time:   time.Duration(duration) * time.Millisecond,
			Artist: t.Artist,
			Title:  t.Title,
			Album:  t.Album,
			Track:  t.TrackNumber,
		}
	}
	g.Songs = songs
	g.Tracks = tracks
	return songs, err
}
