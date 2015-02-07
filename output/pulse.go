// +build linux

package output

import (
	"bytes"
	"encoding/binary"
	"log"

	"github.com/mjibson/mog/_third_party/github.com/mesilliac/pulse-simple"
)

type output struct {
	st *pulse.Stream
}

func get(sampleRate, channels int) (Output, error) {
	o := new(output)
	var err error
	ss := pulse.SampleSpec{pulse.SAMPLE_FLOAT32LE, uint32(sampleRate), uint8(channels)}
	o.st, err = pulse.Playback("mog", "mog", &ss)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (o *output) Push(samples []float32) {
	buf := new(bytes.Buffer)
	for _, s := range samples {
		_ = binary.Write(buf, binary.LittleEndian, s)
	}
	_, err := o.st.Write(buf.Bytes())
	if err != nil {
		log.Println(err)
	}
}

func (o *output) Start() {
}

func (o *output) Stop() {
}
