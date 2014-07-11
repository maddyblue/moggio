package mpa

import (
	"bytes"
	"io"
	"io/ioutil"
	"time"

	"github.com/korandiz/mpa"
	"github.com/mjibson/mog/codec"
)

func init() {
	codec.RegisterCodec("MP3", "\xff\xfa", newSongs)
	codec.RegisterCodec("MP3", "\xff\xfb", newSongs)
	codec.RegisterCodec("MP3", "\xff\xfc", newSongs)
	codec.RegisterCodec("MP3", "\xff\xfd", newSongs)
	codec.RegisterCodec("MP3", "\xff\xfe", newSongs)
	codec.RegisterCodec("MP3", "\xff\xff", newSongs)
}

func newSongs(r io.Reader) ([]codec.Song, error) {
	s, err := newSong(r)
	if err != nil {
		return nil, err
	}
	return []codec.Song{s}, nil
}

type song struct {
	data    []byte
	freq    int
	decoder *mpa.Decoder
	buff    [2][]float32
}

func newSong(r io.Reader) (*song, error) {
	data, err := ioutil.ReadAll(r) // this is stupid
	if err != nil {
		return nil, err
	}

	decoder := mpa.Decoder{Input: bytes.NewBuffer(data)}
	if err := decoder.DecodeFrame(); err != nil {
		return nil, err
	}
	freq := decoder.SamplingFrequency()

	return &song{data: data, freq: freq}, nil
}

func (s *song) Info() codec.SongInfo {
	return codec.SongInfo{
		Time:       time.Duration(1<<63 - 1), // too hard to tell without decoding
		SampleRate: s.freq,
		Channels:   2, // may vary frame to frame
	}
}

func (s *song) Play(n int) (r []float32) {
	if s.decoder == nil {
		s.decoder = &mpa.Decoder{Input: bytes.NewBuffer(s.data)}
	}
	for len(r) < n {
		if len(s.buff[0]) == 0 {
			if err := s.decoder.DecodeFrame(); err != nil {
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

func (s *song) Close() {
	s.decoder, s.buff[0], s.buff[1] = nil, nil, nil
}
