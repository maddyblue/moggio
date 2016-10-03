package webm

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
	codec.RegisterCodec("WEBM", []string{"\u001A\u0045\u00DF\u00A3"}, []string{"webm"}, NewSongs, nil)
}

func NewSongs(rf codec.Reader) (codec.Songs, error) {
	s, err := NewSong(rf)
	if err != nil {
		return nil, err
	}
	return codec.Songs{codec.None: s}, nil
}

func NewSong(rf codec.Reader) (codec.Song, error) {
	f := &WebM{
		Reader: rf,
	}
	return f, nil
}

type WebM struct {
	Reader  codec.Reader
	r       io.ReadCloser
	samples []float32
	info    *codec.SongInfo
	
	w webm.WebM
	wr *webm.Reader
	dec *ADecoder
}

func (w *WebM) Init() (sampleRate, channels int, err error) {
	if w.v == nil {
		r, _, err := w.Reader()
		if err != nil {
			return 0, 0, err
		}
		
		
		
	reader, err := webm.Parse(r, &w.webm)
	if err != nil {
		return 0, 0, err
	}
	w.r = reader
	atrack := s.meta.FindFirstAudioTrack()
	aPackets := make(chan webm.Packet, 1)
	if atrack == nil {
		return 0, 0, errors.New("no audio track")
	}
	s.dec = NewADecoder(ACodec(atrack.CodecID), atrack.CodecPrivate,
			int(atrack.Channels), int(atrack.SamplingFrequency), aPackets)
	}
	go func() { // demuxer
		for pkt := range s.reader.Chan {
				if pkt.TrackNumber  == atrack.TrackNumber {
					aPackets <- pkt
			}
		}
		close(aPackets)
		s.reader.Shutdown()
	}()
	return s, nil

		
		
		
		
		
		
		
		
		vr, err := vorbis.Open(r)
		if err != nil {
			r.Close()
			return 0, 0, err
		}
		w.r = r
		w.v = vr
	}
	return w.w.SampleRate(), w.w.Channels(), nil
}

func (w *WebM) Info() (info codec.SongInfo, err error) {
		return
}

func (w *WebM) Play(n int) ([]float32, error) {
	var err error
	var samples [][]float32
	for len(w.samples) < n && err == nil {
		samples, err = w.w.DecodePacket()
		if len(samples) == 0 {
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
		w.samples = append(w.samples, data...)
	}
	if n > len(w.samples) {
		n = len(w.samples)
	}
	ret := w.samples[:n]
	w.samples = w.samples[n:]
	return ret, err
}

func (w *WebM) Close() {
	if w.r != nil {
		w.r.Close()
		w.r = nil
	}
	w.v = nil
}
