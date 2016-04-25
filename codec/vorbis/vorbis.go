package vorbis

import (
	"bytes"
	"io"
	"time"

	"github.com/dhowden/tag"
	"github.com/jfreymuth/go-vorbis/ogg"
	"github.com/jfreymuth/go-vorbis/ogg/vorbis"
	"github.com/mjibson/moggio/codec"
)

func init() {
	codec.RegisterCodec("VORBIS", []string{"OggS"}, []string{"ogg"}, NewSongs, nil)
}

func NewSongs(rf codec.Reader) (codec.Songs, error) {
	s, err := NewSong(rf)
	if err != nil {
		return nil, err
	}
	return codec.Songs{codec.None: s}, nil
}

func NewSong(rf codec.Reader) (codec.Song, error) {
	f := &Vorbis{
		Reader: rf,
	}
	return f, nil
}

type Vorbis struct {
	Reader  codec.Reader
	r       io.ReadCloser
	v       *vorbis.Vorbis
	samples []float32
	info    *codec.SongInfo
}

func (v *Vorbis) Init() (sampleRate, channels int, err error) {
	if v.v == nil {
		r, _, err := v.Reader()
		if err != nil {
			return 0, 0, err
		}
		vr, err := vorbis.Open(r)
		if err != nil {
			r.Close()
			return 0, 0, err
		}
		v.r = r
		v.v = vr
	}
	return v.v.SampleRate(), v.v.Channels(), nil
}

func (v *Vorbis) Info() (info codec.SongInfo, err error) {
	if v.info != nil {
		return *v.info, nil
	}
	si, _, b, err := v.Reader.Metadata(tag.OGG)
	if err != nil {
		return
	}
	or := ogg.NewReader(bytes.NewReader(b))
	vr, err := vorbis.OpenOgg(or)
	if err != nil {
		return
	}
	l, err := or.Length()
	if err != nil {
		return
	}
	si.Time = time.Duration(l/uint64(vr.SampleRate())) * time.Second
	v.info = si
	return *si, nil
}

func (v *Vorbis) Play(n int) ([]float32, error) {
	var err error
	for len(v.samples) < n && err == nil {
		samples, err := v.v.DecodePacket()
		if err != nil {
			break
		}
		n := len(samples[0])
		c := len(samples)
		data := make([]float32, c*n)
		for i, cs := range samples {
			for j, s := range cs {
				data[j*c+i] = s
			}
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
		v.r = nil
	}
	v.v = nil
}
