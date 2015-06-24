// build cgo

package vorbis

import (
	"bytes"
	"io"

	"github.com/mjibson/mog/_third_party/github.com/mjibson/vorbis-fork"
	"github.com/mjibson/mog/codec"
)

func init() {
	codec.RegisterCodec("VORBIS", "OggS", []string{"ogg"}, NewSongs)
}

func NewSongs(rf codec.Reader) ([]codec.Song, error) {
	s, err := NewSong(rf)
	if err != nil {
		return nil, err
	}
	return []codec.Song{s}, nil
}

func NewSong(rf codec.Reader) (codec.Song, error) {
	f := &Vorbis{
		Reader: rf,
	}
	_, _, err := f.Init()
	return f, err
}

type Vorbis struct {
	Reader  codec.Reader
	r       io.ReadCloser
	initbuf []byte
	v       *vorbis.Vorbis
	samples []float32
}

func (v *Vorbis) Init() (sampleRate, channels int, err error) {
	if v.v == nil {
		r, _, err := v.Reader()
		if err != nil {
			return 0, 0, err
		}
		buf := new(bytes.Buffer)
		defer func() {
			v.initbuf = buf.Bytes()
		}()
		vr, err := vorbis.New(io.TeeReader(r, buf))
		if err != nil {
			r.Close()
			return 0, 0, err
		}
		v.r = r
		v.v = vr
	}
	return v.v.SampleRate, v.v.Channels, nil
}

func (v *Vorbis) Info() (info codec.SongInfo, err error) {
	return
}

func (v *Vorbis) Play(n int) ([]float32, error) {
	var err error
	var data []float32
	for len(v.samples) < n && err == nil {
		data, err = v.v.Decode()
		if err != nil {
			break
		}
		v.samples = append(v.samples, data...)
	}
	if n > len(v.samples) {
		n = len(v.samples)
	}
	ret := v.samples[:n]
	v.samples = v.samples[n:]
	return ret, err
}

func (v *Vorbis) Close() {
	if v.r != nil {
		v.r.Close()
	}
	if v.v != nil {
		v.v.Close()
		v.v = nil
	}
}
