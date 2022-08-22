//go:build linux
// +build linux

package output

import (
	"errors"
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

func (o *output) init() error {
	if o.stream != nil {
		o.stream.Drain()
		o.stream.Close()
		o.stream = nil
	}
	if o.client != nil {
		o.client.Close()
		o.stream = nil
	}
	var err error
	o.client, err = pulse.NewClient(pulse.ClientApplicationName("moggio"))
	if err != nil {
		return fmt.Errorf("pulse client: %w", err)
	}
	var channels pulse.PlaybackOption
	switch o.channels {
	case 1:
		channels = pulse.PlaybackMono
	case 2:
		channels = pulse.PlaybackStereo
	default:
		return errors.New("unsupported channels")
	}
	// The pulse package sometimes gets stuck here waiting on a recv, so allow it
	// to timeout and error (no idea what happens to the abandoned one).
	type newStream struct {
		stream *pulse.PlaybackStream
		err    error
	}
	ch := make(chan newStream, 1)
	go func() {
		stream, err := o.client.NewPlayback(
			pulse.Float32Reader(o.reader),
			channels,
			pulse.PlaybackSampleRate(int(o.sampleRate)),
			pulse.PlaybackLatency(.1),
		)
		ch <- newStream{stream, err}
	}()
	select {
	case newStream := <-ch:
		if newStream.err != nil {
			return fmt.Errorf("pulse stream new: %w", newStream.err)
		}
		o.stream = newStream.stream
		o.samplesChan = make(chan []float32)
		return nil
	case <-time.After(time.Second):
		return errors.New("pulse stream: timeout making new stream")
	}
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
	err := o.init()
	return o, err
}

func (o *output) Push(samples []float32) {
	select {
	case o.samplesChan <- samples:
	case <-time.After(100 * time.Millisecond):
		if o.stream != nil {
			if err := o.stream.Error(); err != nil {
				log.Println("pulse stream error:", err)
			}
		}
		// Restart the stream.
		log.Println("restarting pulse client")
		if err := o.init(); err != nil {
			log.Println("could not restart pulse:", err)
			return
		}
		go func() {
			o.samplesChan <- samples
		}()
		o.stream.Start()
		log.Println("restarted pulse")
	}
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
