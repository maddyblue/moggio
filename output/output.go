package output

type Output interface {
	// Push puts the sample on the output buffer.
	Push(samples []float32)
	// Dispose performs needed closing operations.
	Dispose()
}
