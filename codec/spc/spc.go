// build cgo

package spc

import (
	"io/ioutil"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/mjibson/mog/_third_party/github.com/mjibson/spc"
	"github.com/mjibson/mog/codec"
)

func init() {
	codec.RegisterCodec("SPC", "SNES-SPC", []string{"spc"}, NewSongs)
}

func NewSongs(rf codec.Reader) ([]codec.Song, error) {
	s, err := NewSong(rf)
	if err != nil {
		return nil, err
	}
	return []codec.Song{s}, nil
}

func NewSong(rf codec.Reader) (codec.Song, error) {
	f := &SPC{
		Reader: rf,
	}
	_, _, err := f.Init()
	return f, err
}

type SPC struct {
	Reader codec.Reader
	b      []byte
	s      *spc.SPC
}

func (s *SPC) get() ([]byte, error) {
	if s.b == nil {
		r, _, err := s.Reader()
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		s.b = b
	}
	return s.b, nil
}

func (s *SPC) Init() (sampleRate, channels int, err error) {
	if s.s == nil {
		b, err := s.get()
		if err != nil {
			return 0, 0, err
		}
		ss, err := spc.New(b)
		if err != nil {
			return 0, 0, err
		}
		s.s = ss
	}
	return spc.SampleRate(), 2, nil
}

func (s *SPC) Info() (info codec.SongInfo, err error) {
	b, err := s.get()
	if err != nil {
		return
	}
	info.Title = clean(b[0x2e : 0x2e+32])
	info.Album = clean(b[0x4e : 0x4e+32])
	info.Artist = clean(b[0xb1 : 0xb1+32])
	if i, _ := strconv.Atoi(clean(b[0xa9 : 0xa9+3])); i != 0 {
		info.Time = time.Second * time.Duration(i)
	}
	return
}

func clean(b []byte) string {
	s := strings.TrimSpace(string(b))
	return strings.Split(s, "\x00")[0]
}

func (s *SPC) Play(n int) ([]float32, error) {
	data := make([]int16, n)
	if err := s.s.Play(data); err != nil {
		return nil, err
	}
	ret := make([]float32, n)
	for i, s := range data {
		ret[i] = (float32(s) - math.MinInt16) / (math.MaxInt16 - math.MinInt16)
	}
	return ret, nil
}

func (s *SPC) Close() {
	s.b = nil
	if s.s != nil {
		s.s.Close()
		s.s = nil
	}
}
