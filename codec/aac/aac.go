// +build cgo

package aac

import (
	"fmt"

	"github.com/mjibson/moggio/codec"
	"github.com/nareix/joy4/av"
	"github.com/nareix/joy4/av/avutil"
	"github.com/nareix/joy4/cgo/ffmpeg"
)

func init() {
	codec.RegisterCodec("AAC", []string{"\u00ff\u00f1", "\u00ff\u00f9"}, []string{"aac"}, NewSongs, nil)
}

func NewSongs(rf codec.Reader) (codec.Songs, error) {
	s, err := NewSong(rf)
	if err != nil {
		return nil, err
	}
	return codec.Songs{codec.None: s}, nil
}

func NewSong(rf codec.Reader) (codec.Song, error) {
	f := &AAC{
		Reader: rf,
	}
	return f, nil
}

type AAC struct {
	Reader codec.Reader
	//dm      *mp4.Demuxer
	dmc     av.DemuxCloser
	dec     *ffmpeg.AudioDecoder
	stream  int8
	samples []float32
}

func (v *AAC) Init() (sampleRate, channels int, err error) {
	if v.dec == nil {
		/*
			r, _, err := v.Reader()
			if err != nil {
				return 0, 0, err
			}
			b, err := ioutil.ReadAll(r)
			r.Close()
			if err != nil {
				return 0, 0, err
			}
			fmt.Println("LB", len(b))
			v.dm = mp4.NewDemuxer(bytes.NewReader(b))
			streams, err := v.dm.Streams()
			if err != nil {
				return 0, 0, err
			}
		*/
		file, err := avutil.Open("test.file")
		if err != nil {
			return 0, 0, err
		}
		v.dmc = file
		streams, err := file.Streams()
		if err != nil {
			return 0, 0, err
		}
		for i, s := range streams {
			if s.Type() == av.AAC {
				v.dec, err = ffmpeg.NewAudioDecoder(s.(av.AudioCodecData))
				if err != nil {
					return 0, 0, err
				}
				v.stream = int8(i)
				break
			}
		}
		if v.dec == nil {
			return 0, 0, err
		}
		if err := v.dec.Setup(); err != nil {
			return 0, 0, err
		}
	}
	return v.dec.SampleRate, v.dec.ChannelLayout.Count(), nil
}

func (v *AAC) Info() (info codec.SongInfo, err error) {
	return
}

func (v *AAC) Play(n int) ([]float32, error) {
	var err error
	var pkt av.Packet
	var ok bool
	var frame av.AudioFrame
	for len(v.samples) < n && err == nil {
		println(1)
		pkt, err = v.dmc.ReadPacket()
		fmt.Println("ERR", err, "PKT", pkt)
		if err != nil {
			println(2)
			break
		}
		if pkt.Idx != v.stream {
			continue
		}
		ok, frame, err = v.dec.Decode(pkt.Data)
		if err != nil {
			println(3)
			break
		}
		if ok {
			println(4)
			switch frame.SampleFormat {
			default:
				return nil, fmt.Errorf("unknown sample format: %s", frame.SampleFormat)
			}
		}
		println(5)
	}
	if n > len(v.samples) {
		n = len(v.samples)
	}
	ret := v.samples[:n]
	v.samples = v.samples[n:]
	return ret, err
}

func (v *AAC) Close() {
	if v.dec != nil {
		v.dec.Close()
		v.dec = nil
		v.dmc.Close()
		v.dmc = nil
	}
}
