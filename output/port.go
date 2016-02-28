// +build darwin

package output

import "github.com/helinwang/portaudio"

type port struct {
	st   *portaudio.Stream
	ch   chan []float32
	over []float32
}

func init() {
	portaudio.Initialize()
}

func get(sampleRate, channels int) (Output, error) {
	o := &port{
		ch: make(chan []float32),
	}
	var err error
	o.st, err = portaudio.OpenDefaultStream(0, channels, float64(sampleRate), 1024, o.Fetch)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (p *port) Push(samples []float32) {
	p.ch <- samples
}

// Fetch pulls out samples from the push channel as needed. It takes care
// of the cases where we need or have more or less samples than desired.
func (p *port) Fetch(out []float32) {
	// Write previously saved samples.
	i := copy(out, p.over)
	p.over = p.over[i:]
	for i < len(out) {
		select {
		case s := <-p.ch:
			n := copy(out[i:], s)
			if n < len(s) {
				// Save anything we didn't need this time.
				p.over = s[n:]
			}
			i += n
		default:
			z := make([]float32, len(out)-i)
			copy(out[i:], z)
			return
		}
	}
}

func (p *port) Stop() {
	p.st.Stop()
}

func (p *port) Start() {
	p.st.Start()
}
