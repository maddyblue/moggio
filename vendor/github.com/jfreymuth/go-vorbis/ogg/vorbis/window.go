package vorbis

type windowType struct {
	size, prev, next int
}

func (s *setup) applyWindow(t *windowType, data [][]float32) {
	center := t.size / 2
	prevOffset := t.size/4 - t.prev/4
	nextOffset := t.size/4 - t.next/4
	var prevType, nextType int
	if t.prev == s.blocksize[1] {
		prevType = 1
	}
	if t.next == s.blocksize[1] {
		nextType = 1
	}
	for ch := range data {
		for i := 0; i < prevOffset; i++ {
			data[ch][i] = 0
		}
		for i := 0; i < t.prev/2; i++ {
			data[ch][prevOffset+i] *= s.windows[prevType][i]
		}
		for i := 0; i < t.next/2; i++ {
			data[ch][center+nextOffset+i] *= s.windows[nextType][t.next/2+i]
		}
		for i := t.size - nextOffset; i < t.size; i++ {
			data[ch][i] = 0
		}
	}
}
