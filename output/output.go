package output

type Output interface {
	// Push puts the sample on the output buffer.
	Push(samples []float32)
	Stop()
	Start()
}

var outputs = make(map[config]Output)

type config struct {
	sr, ch int
}

func Get(sampleRate, channels int) (Output, error) {
	c := config{sampleRate, channels}
	if p, ok := outputs[c]; ok {
		p.Start()
		return p, nil
	}
	p, err := get(sampleRate, channels)
	if err != nil {
		return nil, err
	}
	outputs[c] = p
	p.Start()
	return p, nil
}
