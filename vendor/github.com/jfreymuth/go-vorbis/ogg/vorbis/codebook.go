package vorbis

import (
	"math"

	"github.com/jfreymuth/go-vorbis/ogg"
)

const codebookPattern = 0x564342 //"BCV"

type codebook struct {
	dimensions uint32
	entries    huffmanTable
	values     []float32
}

func (c *codebook) ReadFrom(r *ogg.BitReader) error {
	if r.Read32(24) != codebookPattern {
		return ogg.ErrCorruptStream
	}
	c.dimensions = r.Read32(16)
	numEntries := r.Read32(24)
	c.entries = make(huffmanTable, numEntries*2-2)
	ordered := r.ReadBool()
	if !ordered {
		sparse := r.ReadBool()
		for i := uint32(0); i < numEntries; i++ {
			if !sparse || r.ReadBool() {
				c.entries.Put(i, r.Read8(5)+1)
			}
		}
	} else {
		currentEntry := uint32(0)
		currentLength := r.Read8(5) + 1
		for currentEntry < numEntries {
			num := r.Read32(ilog(int(numEntries - currentEntry)))
			for i := currentEntry; i < currentEntry+num; i++ {
				c.entries.Put(i, currentLength)
			}
			currentEntry += num
			currentLength++
		}
	}

	lookupType := r.Read8(4)
	if lookupType == 0 {
		return nil
	}
	if lookupType > 2 {
		return ogg.ErrCorruptStream
	}
	minimumValue := float32Unpack(r.Read32(32))
	deltaValue := float32Unpack(r.Read32(32))
	valueBits := r.Read8(4) + 1
	sequenceP := r.ReadBool()
	var multiplicands []uint32
	if lookupType == 1 {
		multiplicands = make([]uint32, lookup1Values(int(numEntries), c.dimensions))
	} else {
		multiplicands = make([]uint32, int(numEntries)*int(c.dimensions))
	}
	for i := range multiplicands {
		multiplicands[i] = r.Read32(uint(valueBits))
	}
	c.values = make([]float32, numEntries*c.dimensions)
	for entry := 0; entry < int(numEntries); entry++ {
		index := entry * int(c.dimensions)
		if lookupType == 1 {
			last := float32(0)
			indexDivisor := 1
			for i := 0; i < int(c.dimensions); i++ {
				multiplicandOffset := (entry / indexDivisor) % len(multiplicands)
				c.values[index+i] = float32(multiplicands[multiplicandOffset])*deltaValue + minimumValue + last
				if sequenceP {
					last = c.values[index+i]
				}
				indexDivisor *= len(multiplicands)
			}
		} else if lookupType == 2 {
			last := float32(0)
			for i := 0; i < int(c.dimensions); i++ {
				c.values[index+i] = float32(multiplicands[index+i])*deltaValue + minimumValue + last
				if sequenceP {
					last = c.values[index+i]
				}
			}
		}
	}
	return nil
}

func (c *codebook) DecodeScalar(r *ogg.BitReader) uint32 {
	return c.entries.Lookup(r)
}

func (c *codebook) DecodeVector(r *ogg.BitReader) []float32 {
	index := c.entries.Lookup(r) * c.dimensions
	return c.values[index : index+c.dimensions]
}

func ilog(x int) uint {
	var r uint
	for x > 0 {
		r++
		x >>= 1
	}
	return r
}

func lookup1Values(entries int, dim uint32) int {
	return int(math.Floor(math.Pow(float64(entries), 1/float64(dim))))
}

func float32Unpack(x uint32) float32 {
	mantissa := float64(x & 0x1fffff)
	if x&0x80000000 != 0 {
		mantissa = -mantissa
	}
	exponent := (x & 0x7fe00000) >> 21
	return float32(math.Ldexp(mantissa, int(exponent)-788))
}
