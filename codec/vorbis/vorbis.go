// +build cgo

package vorbis

import (
	"io"

	"github.com/dhowden/tag"
	"github.com/mccoyst/vorbis"
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
		vr, err := vorbis.New(r)
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
	if v.info != nil {
		return *v.info, nil
	}
	si, _, b, err := v.Reader.Metadata(tag.OGG)
	if err != nil {
		return
	}
	si.Time, _ = vorbis.Length(b)
	v.info = si
	return *si, nil
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
