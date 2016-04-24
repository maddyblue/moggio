package vorbis

import (
	"math"

	"github.com/jfreymuth/go-vorbis/ogg"
)

type floorData struct {
	floor     floor
	data      []uint32
	noResidue bool
}

func (s *setup) decodePacket(r *ogg.BitReader, prev [][]float32) ([][]float32, [][]float32, error) {
	if r.ReadBool() {
		return nil, nil, ogg.ErrCorruptStream
	}
	modeNumber := r.Read8(ilog(len(s.modes) - 1))
	mode := s.modes[modeNumber]
	// decode window type
	blocktype := mode.blockflag
	longWindow := mode.blockflag == 1
	blocksize := s.blocksize[blocktype]
	spectrumSize := uint32(blocksize / 2)
	windowPrev, windowNext := false, false
	window := windowType{blocksize, blocksize, blocksize}
	if longWindow {
		windowPrev = r.ReadBool()
		windowNext = r.ReadBool()
		if !windowPrev {
			window.prev = s.blocksize[0]
		}
		if !windowNext {
			window.next = s.blocksize[0]
		}
	}

	mapping := &s.mappings[mode.mapping]
	floors := make([]floorData, s.channels)
	residueVectors := make([][]float32, s.channels)
	for ch := range residueVectors {
		residueVectors[ch] = s.residueBuffer[ch][:spectrumSize]
		for i := range residueVectors[ch] {
			residueVectors[ch][i] = 0
		}
	}

	s.decodeFloors(r, floors, mapping, spectrumSize)
	s.decodeResidue(r, residueVectors, mapping, floors, spectrumSize)
	s.inverseCoupling(mapping, residueVectors)
	s.applyFloor(floors, residueVectors)

	// inverse MDCT
	out := make([][]float32, s.channels)
	for ch := range out {
		out[ch] = make([]float32, blocksize)
		imdct(s.lookup[blocktype], residueVectors[ch], out[ch])
	}

	// apply window and overlap
	s.applyWindow(&window, out)
	center := blocksize / 2
	offset := s.blocksize[1]/4 - s.blocksize[0]/4
	final := make([][]float32, s.channels)
	next := make([][]float32, s.channels)
	if longWindow {
		for ch := range out {
			start := 0
			end := blocksize
			if !windowPrev {
				start += offset
			}
			if !windowNext {
				end -= offset
			}
			final[ch], next[ch] = out[ch][start:center], out[ch][center:end]
			if prev != nil {
				for i := range prev[ch] {
					final[ch][i] += prev[ch][i]
				}
			}
		}
	} else /*short window*/ {
		for ch := range out {
			if prev != nil && len(prev[ch]) > center {
				final[ch], next[ch] = prev[ch], out[ch][center:]
				for i := offset; i < len(final[ch]); i++ {
					final[ch][i] += out[ch][i-offset]
				}
			} else {
				final[ch], next[ch] = out[ch][:center], out[ch][center:]
				if prev != nil {
					for i := range final[ch] {
						final[ch][i] += prev[ch][i]
					}
				}
			}
		}
	}
	return final, next, nil
}

func (s *setup) decodeFloors(r *ogg.BitReader, floors []floorData, mapping *mapping, n uint32) {
	for ch := range floors {
		floor := s.floors[mapping.submaps[mapping.mux[ch]].floor]
		data := floor.Decode(r, s.codebooks, n)
		floors[ch] = floorData{floor, data, data == nil}
	}

	for i := 0; i < int(mapping.couplingSteps); i++ {
		if !floors[mapping.magnitude[i]].noResidue || !floors[mapping.angle[i]].noResidue {
			floors[mapping.magnitude[i]].noResidue = false
			floors[mapping.angle[i]].noResidue = false
		}
	}

}

func (s *setup) decodeResidue(r *ogg.BitReader, out [][]float32, mapping *mapping, floors []floorData, n uint32) {
	for i := range mapping.submaps {
		doNotDecode := make([]bool, 0, len(out))
		tmp := make([][]float32, 0, len(out))
		for j := 0; j < s.channels; j++ {
			if mapping.mux[j] == uint8(i) {
				doNotDecode = append(doNotDecode, floors[j].noResidue)
				tmp = append(tmp, out[j])
			}
		}
		s.residues[mapping.submaps[i].residue].Decode(r, doNotDecode, n, s.codebooks, tmp)
	}
}

func (s *setup) inverseCoupling(mapping *mapping, residueVectors [][]float32) {
	for i := mapping.couplingSteps; i > 0; i-- {
		magnitudeVector := residueVectors[mapping.magnitude[i-1]]
		angleVector := residueVectors[mapping.angle[i-1]]
		for j := range magnitudeVector {
			m := magnitudeVector[j]
			a := angleVector[j]
			if m > 0 {
				if a > 0 {
					m, a = m, m-a
				} else {
					a, m = m, m+a
				}
			} else {
				if a > 0 {
					m, a = m, m+a
				} else {
					a, m = m, m-a
				}
			}
			magnitudeVector[j] = m
			angleVector[j] = a
		}
	}
}

func (s *setup) applyFloor(floors []floorData, residueVectors [][]float32) {
	for ch := range residueVectors {
		if floors[ch].data != nil {
			floors[ch].floor.Apply(residueVectors[ch], floors[ch].data)
		}
	}
}

func makeWindow(size int) []float32 {
	window := make([]float32, size)
	for i := range window {
		window[i] = windowFunc((float32(i) + .5) / float32(size/2) * math.Pi / 2)
	}
	return window
}

func windowFunc(x float32) float32 {
	sinx := math.Sin(float64(x))
	return float32(math.Sin(math.Pi / 2 * sinx * sinx))
}
