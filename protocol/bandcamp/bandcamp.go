package bandcamp

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/mjibson/moggio/codec"
	"github.com/mjibson/moggio/codec/mp3"
	"github.com/mjibson/moggio/protocol"
	"golang.org/x/oauth2"
)

func init() {
	protocol.Register("bandcamp", []string{"URL"}, New, reflect.TypeOf(&Bandcamp{}))
	gob.Register(new(Bandcamp))
}

func New(params []string, token *oauth2.Token) (protocol.Instance, error) {
	if len(params) != 1 {
		return nil, fmt.Errorf("expected one parameter")
	}
	b := &Bandcamp{
		URL: params[0],
	}
	if _, err := b.Refresh(); err != nil {
		return nil, err
	}
	return b, nil
}

type Bandcamp struct {
	URL    string
	Songs  protocol.SongList
	Tracks map[codec.ID]*track
}

func (b *Bandcamp) Key() string {
	return b.URL
}

func (b *Bandcamp) Info(id codec.ID) (*codec.SongInfo, error) {
	t := b.Songs[id]
	if t == nil {
		return nil, fmt.Errorf("could not find %v", id)
	}
	return t, nil
}

func (b *Bandcamp) GetSong(id codec.ID) (codec.Song, error) {
	t := b.Tracks[id]
	if t == nil {
		return nil, fmt.Errorf("missing %v", id)
	}
	return mp3.NewSong(func() (io.ReadCloser, int64, error) {
		log.Println("BANDCAMP", id)
		res, err := http.Get(t.File.Mp3_128)
		if err != nil {
			return nil, 0, err
		}
		return res.Body, 0, nil
	})
}

func (b *Bandcamp) List() (protocol.SongList, error) {
	if len(b.Songs) == 0 {
		return b.Refresh()
	}
	return b.Songs, nil
}

type track struct {
	Duration float64 `json:"duration"`
	File     struct {
		Mp3_128 string `json:"mp3-128"`
	} `json:"file"`
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	TrackNum int    `json:"track_num"`
}

type album struct {
	Artist string `json:"artist"`
	Title  string `json:"title"`
}

func (b *Bandcamp) Refresh() (protocol.SongList, error) {
	resp, err := http.Get(b.URL)
	if err != nil {
		return nil, err
	}
	y, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	var tracks []*track
	var a album
	{
		const start = "    trackinfo : ["
		i := bytes.Index(y, []byte(start))
		if i < 0 {
			return nil, fmt.Errorf("not a bandcamp site: %v", b.URL)
		}
		e := bytes.Index(y[i:], []byte("],\n"))
		if e < 0 {
			return nil, fmt.Errorf("not a bandcamp site: %v", b.URL)
		}
		if err := json.Unmarshal(y[i+len(start)-1:e+1+i], &tracks); err != nil {
			return nil, err
		}
	}
	{
		const start = "    current: {"
		i := bytes.Index(y, []byte(start))
		if i < 0 {
			return nil, fmt.Errorf("not a bandcamp site: %v", b.URL)
		}
		e := bytes.Index(y[i:], []byte("},\n"))
		if e < 0 {
			return nil, fmt.Errorf("not a bandcamp site: %v", b.URL)
		}
		if err := json.Unmarshal(y[i+len(start)-1:e+1+i], &a); err != nil {
			return nil, err
		}
	}
	var artURL string
	func() {
		const start = `    artThumbURL: "`
		i := bytes.Index(y, []byte(start))
		if i < 0 {
			return
		}
		e := bytes.Index(y[i:], []byte("\",\n"))
		if e < 0 {
			return
		}
		if err := json.Unmarshal(y[i+len(start)-1:e+1+i], &artURL); err != nil {
			return
		}
	}()
	tracklist := make(map[codec.ID]*track)
	songs := make(protocol.SongList)
	for _, t := range tracks {
		if t.File.Mp3_128 == "" {
			continue
		}
		if strings.HasPrefix(t.File.Mp3_128, "//") {
			t.File.Mp3_128 = "https:" + t.File.Mp3_128
		}
		id := codec.Int64(t.ID)
		tracklist[id] = t
		songs[id] = &codec.SongInfo{
			Time:     time.Duration(t.Duration) * time.Second,
			Album:    a.Title,
			Artist:   a.Artist,
			Title:    t.Title,
			Track:    float64(t.TrackNum),
			ImageURL: artURL,
		}
	}
	b.Songs = songs
	b.Tracks = tracklist
	return songs, err
}
