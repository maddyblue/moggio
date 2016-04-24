package vorbis

import "math"

//slow implementation of the inverse MDCT
//might be useful for testing

func imdctSlow(in []float32) []float32 {
	n := len(in)
	fn := float32(n)
	out := make([]float32, 2*n)
	for i := range out {
		var sum float32
		fi := float32(i)
		for k := 0; k < n; k++ {
			fk := float32(k)
			sum += in[k] * float32(math.Cos(float64((math.Pi/fn)*(fi+.5+fn/2)*(fk+.5))))
		}
		out[i] = sum
	}
	return out
}
