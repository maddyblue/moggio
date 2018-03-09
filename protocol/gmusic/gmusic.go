package gmusic

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"
	"time"

	"github.com/mjibson/gmusic"
	"github.com/mjibson/moggio/codec"
	"github.com/mjibson/moggio/codec/mp3"
	"github.com/mjibson/moggio/protocol"
	"golang.org/x/oauth2"
)

func init() {
	gob.Register(new(GMusic))
	protocol.Register("gmusic", []string{"username", "password"}, New, reflect.TypeOf(&GMusic{}))
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
		Name:   params[0],
	}, nil
}

func (g *GMusic) Info(id codec.ID) (*codec.SongInfo, error) {
	s := g.Songs[id]
	if s == nil {
		return nil, fmt.Errorf("could not find %v", id)
	}
	return s, nil
}

type GMusic struct {
	GMusic *gmusic.GMusic
	Name   string
	Tracks map[codec.ID]*gmusic.Track
	Songs  protocol.SongList
}

func (g *GMusic) Key() string {
	return g.Name
}

func (g *GMusic) List() (protocol.SongList, error) {
	if len(g.Songs) == 0 {
		return g.Refresh()
	}
	return g.Songs, nil
}

func (g *GMusic) GetSong(id codec.ID) (codec.Song, error) {
	f := g.Tracks[id]
	if f == nil {
		return nil, fmt.Errorf("missing %v", id)
	}
	return mp3.NewSong(func() (io.ReadCloser, int64, error) {
		log.Println("GMUSIC", id)
		r, err := g.GMusic.GetStream(string(id))
		if err != nil {
			return nil, 0, err
		}
		size, _ := strconv.ParseInt(f.EstimatedSize, 10, 64)
		return r.Body, size, nil
	})
}

func (g *GMusic) Refresh() (protocol.SongList, error) {
	tracks := make(map[codec.ID]*gmusic.Track)
	songs := make(protocol.SongList)
	log.Println("get gmusic tracks")
	trackList, err := g.GMusic.ListTracks()
	if err != nil {
		return nil, err
	}
	log.Println("got gmusic tracks", len(trackList))
	for _, t := range trackList {
		tracks[codec.ID(t.ID)] = t
		duration, _ := strconv.Atoi(t.DurationMillis)
		si := &codec.SongInfo{
			Time:   time.Duration(duration) * time.Millisecond,
			Artist: t.AlbumArtist,
			Title:  t.Title,
			Album:  t.Album,
			Track:  t.TrackNumber,
		}
		if len(t.AlbumArtRef) != 0 {
			si.ImageURL = t.AlbumArtRef[0].URL
		}
		songs[codec.ID(t.ID)] = si
	}
	g.Songs = songs
	g.Tracks = tracks
	return songs, err
}
