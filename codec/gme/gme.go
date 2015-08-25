// +build cgo

package gme

import (
	"io"
	"io/ioutil"
	"math"
	"strconv"

	"github.com/mjibson/gme"
	"github.com/mjibson/mog/codec"
)

func init() {
	codec.RegisterCodec("SPC", []string{"SNES-SPC"}, []string{"spc"}, NewSongs, GetSong)
	codec.RegisterCodec("NSF", []string{"NESM\u001a"}, []string{"nsf"}, NewSongs, GetSong)
	codec.RegisterCodec("NSFE", []string{"NSFE"}, []string{"nsfe"}, NewSongs, GetSong)
}

const (
	defaultChannels   = 2
	defaultSampleRate = 44100
)

func GetSong(rf codec.Reader, id codec.ID) (codec.Song, error) {
	i, err := strconv.Atoi(string(id))
	if err != nil {
		return nil, err
	}
	return &Track{
		r:     &reader{r: rf},
		track: i,
	}, nil
}

func NewSongs(rf codec.Reader) (codec.Songs, error) {
	d := reader{
		r: rf,
	}
	b, err := d.get()
	if err != nil {
		return nil, err
	}
	gg, err := gme.New(b, defaultSampleRate)
	if err != nil {
		return nil, err
	}
	songs := make(codec.Songs)
	for i, tr := 0, gg.Tracks(); i < tr; i++ {
		songs[codec.Int(i)] = &Track{
			r:     &d,
			track: i,
		}
	}
	return songs, nil
}

type Track struct {
	r     *reader
	g     *gme.GME
	track int
}

func (t *Track) Info() (codec.SongInfo, error) {
	var si codec.SongInfo
	b, err := t.r.get()
	if err != nil {
		return si, err
	}
	g, err := gme.New(b, gme.InfoOnly)
	if err != nil {
		return si, err
	}
	info, err := g.Track(t.track)
	g.Close()
	return codec.SongInfo{
		Time:   info.PlayLength + gme.FadeLength,
		Artist: info.Author,
		Title:  info.Song,
		Album:  info.Game,
		Track:  float64(t.track),
	}, err
}

func (t *Track) Init() (sampleRate, channels int, err error) {
	b, err := t.r.get()
	if err != nil {
		return 0, 0, err
	}
	g, err := gme.New(b, defaultSampleRate)
	if err != nil {
		return 0, 0, err
	}
	if err := g.Start(t.track); err != nil {
		return 0, 0, nil
	}
	t.g = g
	return defaultSampleRate, defaultChannels, nil
}

func (t *Track) Close() {
	if t.g != nil {
		t.g.Close()
		t.g = nil
	}
}

type reader struct {
	r codec.Reader
	b []byte
}

func (d *reader) get() ([]byte, error) {
	if d.b == nil {
		r, _, err := d.r()
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		d.b = b
	}
	return d.b, nil
}

func (t *Track) Play(n int) ([]float32, error) {
	data := make([]int16, n)
	err := t.g.Play(data)
	if err != nil && err != io.EOF {
		return nil, err
	}
	ret := make([]float32, n)
	for i, s := range data {
		ret[i] = (float32(s) - math.MinInt16) / (math.MaxInt16 - math.MinInt16)
	}
	return ret, err
}
