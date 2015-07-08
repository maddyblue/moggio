package mpa

import (
	"bytes"
	"io"
	"io/ioutil"

	"github.com/mjibson/mog/_third_party/github.com/mjibson/go-dsp/wav"
	"github.com/mjibson/mog/codec"
)

func init() {
	codec.RegisterCodec("WAV", []string{"RIFF????WAVE"}, []string{"wav"}, New, nil)
}

func New(rf codec.Reader) (codec.Songs, error) {
	w := Wav{
		Reader: rf,
	}
	return codec.Songs{codec.None: &w}, nil
}

type Wav struct {
	Reader  codec.Reader
	r       io.ReadCloser
	initbuf []byte
	w       *wav.Wav
}

func (w *Wav) Init() (sampleRate, channels int, err error) {
	if w.w == nil {
		r, _, err := w.Reader()
		if err != nil {
			return 0, 0, err
		}
		buf := new(bytes.Buffer)
		defer func() {
			w.initbuf = buf.Bytes()
		}()
		wr, err := wav.New(io.TeeReader(r, buf))
		if err != nil {
			r.Close()
			return 0, 0, err
		}
		w.r = r
		w.w = wr
	}
	return int(w.w.SampleRate), int(w.w.NumChannels), nil
}

func (w *Wav) Info() (info codec.SongInfo, err error) {
	var r io.ReadCloser
	if len(w.initbuf) != 0 {
		r = ioutil.NopCloser(bytes.NewBuffer(w.initbuf))
	}
	if r == nil {
		r, _, err = w.Reader()
		if err != nil {
			return
		}
	}
	wv, err := wav.New(r)
	r.Close()
	if err != nil {
		return
	}
	return codec.SongInfo{
		Time: wv.Duration,
	}, nil
}

func (w *Wav) Play(n int) ([]float32, error) {
	return w.w.ReadFloats(n)
}

func (w *Wav) Close() {
	if w.r != nil {
		w.r.Close()
	}
	if w.w != nil {
		w.w = nil
	}
}
