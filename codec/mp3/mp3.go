package mp3

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"time"

	"github.com/dhowden/tag"
	"github.com/hajimehoshi/go-mp3"
	"github.com/korandiz/mpseek"
	"github.com/mjibson/moggio/codec"
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
	decoder *mp3.Decoder
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
	s.decoder, err = mp3.NewDecoder(r)
	if err != nil {
		r.Close()
		return 0, 0, err
	}
	s.r = r
	return s.decoder.SampleRate(), 2, nil
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

func (s *Song) Play(n int) (r []float32, err error) {
	var samples [2]int16
	for len(r) < n {
		if err := binary.Read(s.decoder, binary.LittleEndian, &samples); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		r = append(r,
			(float32(samples[0]))/(math.MaxInt16-math.MinInt16),
			(float32(samples[1]))/(math.MaxInt16-math.MinInt16),
		)
	}
	return
}

func (s *Song) Close() {
	if s.r != nil {
		s.r.Close()
	}
	s.decoder, s.r = nil, nil
}
