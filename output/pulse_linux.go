package output

import (
	"fmt"

	"github.com/mjibson/pulsego"
)

type pulse struct {
	pa  *pulsego.PulseMainLoop
	ctx *pulsego.PulseContext
	st  *pulsego.PulseStream
}

func NewPulse(sampleRate, channels int) (Output, error) {
	p := pulse{
		pa: pulsego.NewPulseMainLoop(),
	}
	p.pa.Start()

	p.ctx = p.pa.NewContext("default", 0)
	if p.ctx == nil {
		p.Dispose()
		return nil, fmt.Errorf("output: pulse: failed to create new context")
	}
	p.st = p.ctx.NewStream("default", &pulsego.PulseSampleSpec{
		Format: pulsego.SAMPLE_FLOAT32LE, Rate: sampleRate, Channels: channels})
	if p.st == nil {
		p.Dispose()
		return nil, fmt.Errorf("output: pulse: failed to create new stream")
	}
	p.st.ConnectToSink()
	return &p, nil
}

func (p *pulse) Dispose() {
	if p.pa != nil {
		p.pa.Dispose()
		p.pa = nil
	}
	if p.ctx != nil {
		p.ctx.Dispose()
		p.ctx = nil
	}
	if p.st != nil {
		p.st.Dispose()
		p.st = nil
	}
}

func (p *pulse) Push(s []float32) {
	p.st.Write(s, pulsego.SEEK_RELATIVE)
}
