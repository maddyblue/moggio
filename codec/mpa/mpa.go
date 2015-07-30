package mpa

import (
	"bytes"
	"io"
	"math"
	"time"

	"github.com/mjibson/mog/_third_party/github.com/dhowden/tag"
	"github.com/mjibson/mog/_third_party/github.com/korandiz/mpa"
	"github.com/mjibson/mog/_third_party/github.com/korandiz/mpseek"
	"github.com/mjibson/mog/codec"
)

func init() {
	codec.RegisterCodec("MP3",
		[]string{"\xff\xfb", string([]byte{49, 44, 33})},
		[]string{"mp3"},
		NewSongs,
		nil,
	)
}

func NewSongs(rf codec.Reader) (codec.Songs, error) {
	s, err := NewSong(rf)
	if err != nil {
		return nil, err
	}
	return codec.Songs{codec.None: s}, nil
}

type Song struct {
	Reader  codec.Reader
	r       io.ReadCloser
	decoder *mpa.Decoder
	buff    [2][]float32
	info    *codec.SongInfo
}

func NewSong(rf codec.Reader) (*Song, error) {
	s := &Song{Reader: rf}
	return s, nil
}

func (s *Song) Init() (sampleRate, channels int, err error) {
	r, _, err := s.Reader()
	if err != nil {
		return 0, 0, err
	}
	s.decoder = &mpa.Decoder{Input: r}
	s.r = r
	if err := s.decode(); err != nil {
		r.Close()
		return 0, 0, err
	}
	return s.decoder.SamplingFrequency(), s.decoder.NChannels(), nil
}

func (s *Song) Info() (info codec.SongInfo, err error) {
	if s.info != nil {
		return *s.info, nil
	}
	si, _, b, err := s.Reader.Metadata(tag.MP3)
	if err != nil {
		return
	}
	table, err := mpseek.CreateTable(bytes.NewReader(b), math.MaxFloat64)
	if err != nil {
		return
	}
	si.Time = time.Duration(table.Length()) * time.Second
	s.info = si
	return *si, nil
}

func (s *Song) decode() error {
	s.buff[0] = nil
	s.buff[1] = nil
	for {
		if err := s.decoder.DecodeFrame(); err != nil {
			switch err.(type) {
			case mpa.MalformedStream:
				continue
			}
			return err
		}
		break
	}
	for i := 0; i < 2; i++ {
		s.buff[i] = make([]float32, s.decoder.NSamples())
		s.decoder.ReadSamples(i, s.buff[i])
	}
	return nil
}

func (s *Song) Play(n int) (r []float32, err error) {
	for len(r) < n {
		if len(s.buff[0]) == 0 {
			if err = s.decode(); err != nil {
				return
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
