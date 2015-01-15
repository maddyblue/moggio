// +build linux

package output

import (
	"bytes"
	"encoding/binary"
	"log"

	"code.google.com/p/portaudio-go/portaudio"
	"github.com/mesilliac/pulse-simple"
)

var ports = make(map[config]*port)

type config struct {
	sr, ch int
}

type port struct {
	st *pulse.Stream
}

func Get(sampleRate, channels int) (Output, error) {
	c := config{
		sr: sampleRate,
		ch: channels,
	}
	o := ports[c]
	if o == nil {
		portaudio.Initialize()
		o = new(port)
		var err error
		ss := pulse.SampleSpec{pulse.SAMPLE_FLOAT32LE, uint32(sampleRate), uint8(channels)}
		o.st, err = pulse.Playback("mog", "mog", &ss)
		if err != nil {
			return nil, err
		}
		ports[c] = o
	}
	return o, nil
}

func (p *port) Push(samples []float32) {
	buf := new(bytes.Buffer)
	for _, s := range samples {
		_ = binary.Write(buf, binary.LittleEndian, s)
	}
	_, err := p.st.Write(buf.Bytes())
	if err != nil {
		log.Println(err)
	}
}
