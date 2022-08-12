//go:build linux
// +build linux

package output

import (
	"bytes"
	"encoding/binary"
	"log"

	"github.com/mesilliac/pulse-simple"
)

type output struct {
	st         *pulse.Stream
	sampleRate uint32
	channels   uint8
}

func (o *output) init() {
	if o.st != nil {
		o.st.Drain()
		o.st.Free()
		o.st = nil
	}
	ss := pulse.SampleSpec{pulse.SAMPLE_FLOAT32LE, o.sampleRate, o.channels}
	var err error
	o.st, err = pulse.Playback("moggio", "moggio", &ss)
	if err != nil {
		log.Println(err)
		return
	}
}

func get(sampleRate, channels int) (Output, error) {
	o := new(output)
	o.sampleRate = uint32(sampleRate)
	o.channels = uint8(channels)
	o.init()
	return o, nil
}

func (o *output) Push(samples []float32) {
	buf := new(bytes.Buffer)
	for _, s := range samples {
		_ = binary.Write(buf, binary.LittleEndian, s)
	}
	if _, err := o.st.Write(buf.Bytes()); err != nil {
		log.Println(err)
		o.init()
	}
}

func (o *output) Start() {
}

func (o *output) Stop() {
}
