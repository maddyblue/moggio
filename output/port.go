package output

import (
	"log"
	"runtime/debug"

	"code.google.com/p/portaudio-go/portaudio"
)

var (
	portInitCount = 0
)

type port struct {
	st   *portaudio.Stream
	ch   chan []float32
	over []float32
}

func NewPort(sampleRate, channels int) (Output, error) {
	log.Println("NEW PORT", portInitCount)
	debug.PrintStack()
	// todo: fix race condition
	if portInitCount == 0 {
		portaudio.Initialize()
	}
	portInitCount++

	p := port{
		ch: make(chan []float32),
	}
	var err error
	p.st, err = portaudio.OpenDefaultStream(0, channels, float64(sampleRate), 1024, p.Fetch)
	if err != nil {
		p.Dispose()
		return nil, err
	}
	if err := p.st.Start(); err != nil {
		p.Dispose()
		return nil, err
	}
	return &p, nil
}

func (p *port) Push(samples []float32) {
	log.Println("push", len(samples), p.ch)
	p.ch <- samples
	log.Println("pushed", p.ch)
}

// Fetch pulls out samples from the push channel as needed. It takes care
// of the cases where we need or have more or less samples than desired.
func (p *port) Fetch(out []float32) {
	log.Println("fetch", len(out))
	// Write previously saved samples.
	i := copy(out, p.over)
	p.over = p.over[i:]
	for i < len(out) {
		log.Println("fetch ch recv", p.ch)
		s := <-p.ch
		log.Println("fetch ch got")
		n := copy(out[i:], s)
		if n < len(s) {
			// Save anything we didn't need this time.
			p.over = s[n:]
		}
		i += n
	}
	log.Println("fetch done")
}

func (p *port) Dispose() {
	portInitCount--
	if portInitCount == 0 {
		portaudio.Terminate()
	}
	if p.st != nil {
		_ = p.st.Stop() // ignore error
		p.st.Close()
	}
}
