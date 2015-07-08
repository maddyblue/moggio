package flac

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"strconv"
	"time"

	"github.com/mjibson/mog/_third_party/gopkg.in/mewkiz/flac.v1"
	"github.com/mjibson/mog/_third_party/gopkg.in/mewkiz/flac.v1/frame"
	"github.com/mjibson/mog/_third_party/gopkg.in/mewkiz/flac.v1/meta"
	"github.com/mjibson/mog/codec"
)

func init() {
	codec.RegisterCodec("FLAC", []string{"fLaC"}, []string{"flac"}, New, nil)
}

func New(rf codec.Reader) (codec.Songs, error) {
	f := Flac{
		Reader: rf,
	}
	return codec.Songs{codec.None: &f}, nil
}

type Flac struct {
	Reader  codec.Reader
	r       io.ReadCloser
	initbuf []byte
	f       *flac.Stream
	samples []int32
}

func (f *Flac) Init() (sampleRate, channels int, err error) {
	if f.f == nil {
		r, _, err := f.Reader()
		if err != nil {
			return 0, 0, err
		}
		buf := new(bytes.Buffer)
		defer func() {
			f.initbuf = buf.Bytes()
		}()
		fr, err := flac.Parse(io.TeeReader(r, buf))
		if err != nil {
			r.Close()
			return 0, 0, err
		}
		f.r = r
		f.f = fr
	}
	return int(f.f.Info.SampleRate), int(f.f.Info.NChannels), nil
}

func (f *Flac) Info() (info codec.SongInfo, err error) {
	var r io.ReadCloser
	if len(f.initbuf) != 0 {
		r = ioutil.NopCloser(bytes.NewBuffer(f.initbuf))
	}
	if r == nil {
		r, _, err = f.Reader()
		if err != nil {
			return
		}
	}
	fv, err := flac.Parse(r)
	r.Close()
	if err != nil {
		return
	}
	si := codec.SongInfo{
		Time: time.Duration(fv.Info.NSamples) / time.Duration(fv.Info.SampleRate) * time.Second,
	}
	for _, b := range fv.Blocks {
		switch v := b.Body.(type) {
		case *meta.VorbisComment:
			for _, tag := range v.Tags {
				switch tag[0] {
				case "TITLE":
					si.Title = tag[1]
				case "ARTIST":
					si.Artist = tag[1]
				case "ALBUM":
					si.Album = tag[1]
				case "TRACKNUMBER":
					n, _ := strconv.Atoi(tag[1])
					si.Track = float64(n)
				}
			}
		case *meta.Picture:
			if v.MIME == "-->" {
				si.ImageURL = string(v.Data)
				break
			}
			si.ImageURL = fmt.Sprintf("data:%s;base64,%s", v.MIME, base64.StdEncoding.EncodeToString(v.Data))
		}
	}
	return si, nil
}

func (f *Flac) Play(n int) ([]float32, error) {
	var err error
	var frame *frame.Frame
	for len(f.samples) < n && err == nil {
		frame, err = f.f.ParseNext()
		if err != nil {
			break
		}
		for i := 0; i < int(frame.BlockSize); i++ {
			for _, sf := range frame.Subframes {
				f.samples = append(f.samples, sf.Samples[i])
			}
		}
	}
	if n > len(f.samples) {
		n = len(f.samples)
	}
	ret := make([]float32, n)
	for i, s := range f.samples[:n] {
		ret[i] = (float32(s) - math.MinInt16) / (math.MaxInt16 - math.MinInt16)
	}
	f.samples = f.samples[n:]
	return ret, err
}

func (f *Flac) Close() {
	if f.r != nil {
		f.r.Close()
	}
	if f.f != nil {
		f.f = nil
	}
}
