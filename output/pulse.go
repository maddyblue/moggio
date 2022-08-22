//go:build linux
// +build linux

package output

import (
	"fmt"
	"log"
	"time"

	"github.com/jfreymuth/pulse"
)

type output struct {
	client      *pulse.Client
	stream      *pulse.PlaybackStream
	samplesChan chan []float32
	samplesBuf  []float32
	sampleRate  uint32
	channels    uint8
}

func (o *output) init() {
	if o.client != nil {
		o.stream.Drain()
		o.stream.Close()
		o.stream = nil
	}
	var err error
	o.client, err = pulse.NewClient(pulse.ClientApplicationName("moggio"))
	if err != nil {
		log.Println(err)
		return
	}
	var channels pulse.PlaybackOption
	switch o.channels {
	case 1:
		channels = pulse.PlaybackMono
	case 2:
		channels = pulse.PlaybackStereo
	default:
		log.Println("unsupported channels")
		return
	}
	o.stream, err = o.client.NewPlayback(
		pulse.Float32Reader(o.reader),
		channels,
		pulse.PlaybackSampleRate(int(o.sampleRate)),
		pulse.PlaybackLatency(.1),
	)
	if err != nil {
		log.Println(err)
		return
	}
	o.samplesChan = make(chan []float32)
}

func (o *output) reader(out []float32) (int, error) {
	if len(o.samplesBuf) > 0 {
		n := copy(out, o.samplesBuf)
		o.samplesBuf = o.samplesBuf[n:]
		return n, nil
	}
	select {
	case samples := <-o.samplesChan:
		n := copy(out, samples)
		o.samplesBuf = samples[n:]
		return n, nil
	case <-time.After(100 * time.Millisecond):
		return 0, nil
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
	o.samplesChan <- samples
}

func (o *output) Start() {
	// This waits until the buffer is full, so fill it in the background.
	samples := make([]float32, o.stream.BufferSize())
	go o.Push(samples)
	o.stream.Start()
}

func (o *output) Stop() {
	o.stream.Stop()
}
