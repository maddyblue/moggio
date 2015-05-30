package mpa

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/mjibson/mog/_third_party/github.com/korandiz/mpa"
	"github.com/mjibson/mog/_third_party/github.com/mjibson/id3"
	"github.com/mjibson/mog/codec"
)

func init() {
	exts := []string{"mp3"}
	codec.RegisterCodec("MP3", "\xff\xfa", exts, NewSongs)
	codec.RegisterCodec("MP3", "\xff\xfb", exts, NewSongs)
	codec.RegisterCodec("MP3", "\xff\xfc", exts, NewSongs)
	codec.RegisterCodec("MP3", "\xff\xfd", exts, NewSongs)
	codec.RegisterCodec("MP3", "\xff\xfe", exts, NewSongs)
	codec.RegisterCodec("MP3", "\xff\xff", exts, NewSongs)
}

func NewSongs(rf codec.Reader) ([]codec.Song, error) {
	s, err := NewSong(rf)
	if err != nil {
		return nil, err
	}
	return []codec.Song{s}, nil
}

type Song struct {
	Reader  codec.Reader
	r       io.ReadCloser
	decoder *mpa.Decoder
	initbuf []byte
	buff    [2][]float32
}

func NewSong(rf codec.Reader) (*Song, error) {
	s := &Song{Reader: rf}
	_, _, err := s.Init()
	return s, err
}

func (s *Song) Init() (sampleRate, channels int, err error) {
	if s.decoder == nil {
		r, _, err := s.Reader()
		if err != nil {
			return 0, 0, err
		}
		buf := new(bytes.Buffer)
		defer func() {
			s.initbuf = buf.Bytes()
		}()
		s.decoder = &mpa.Decoder{Input: io.TeeReader(r, buf)}
		s.r = r
		for {
			if err := s.decoder.DecodeFrame(); err != nil {
				switch err.(type) {
				case mpa.MalformedStream:
					continue
				}
				r.Close()
				return 0, 0, err
			}
			break
		}
	}
	return s.decoder.SamplingFrequency(), s.decoder.NChannels(), nil
}

func (s *Song) Info() (info codec.SongInfo, err error) {
	var r io.ReadCloser
	if len(s.initbuf) != 0 {
		r = ioutil.NopCloser(bytes.NewBuffer(s.initbuf))
	}
	if r == nil {
		r, _, err = s.Reader()
		if err != nil {
			return
		}
	}
	f := id3.Read(r)
	r.Close()
	if f == nil {
		err = fmt.Errorf("could not read id3 data")
		return
	}
	track, _ := strconv.ParseFloat(f.Track, 64)
	dur, _ := strconv.Atoi(f.Length)
	si := codec.SongInfo{
		Artist: f.Artist,
		Title:  f.Name,
		Album:  f.Album,
		Track:  track,
		Time:   time.Duration(dur) * time.Millisecond,
	}
	if f.Image != nil {
		si.ImageURL = f.Image.DataURL()
	}
	return si, nil
}

func (s *Song) Play(n int) (r []float32, err error) {
	for len(r) < n {
		if len(s.buff[0]) == 0 {
			err = s.decoder.DecodeFrame()
			if err != nil {
				switch err.(type) {
				case mpa.MalformedStream:
					continue
				}
				return
			}
			for i := 0; i < 2; i++ {
				s.buff[i] = make([]float32, s.decoder.NSamples())
				s.decoder.ReadSamples(i, s.buff[i])
			}
		}
		for len(s.buff[0]) > 0 && len(r) < n {
			r = append(r, s.buff[0][0], s.buff[1][0])
			s.buff[0], s.buff[1] = s.buff[0][1:], s.buff[1][1:]
		}
	}
	return
}

func (s *Song) Close() {
	if s.r != nil {
		s.r.Close()
	}
	s.decoder, s.buff[0], s.buff[1], s.r = nil, nil, nil, nil
}
