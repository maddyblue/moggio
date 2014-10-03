package output

type Output interface {
	// Push puts the sample on the output buffer.
	Push(samples []float32)
}
