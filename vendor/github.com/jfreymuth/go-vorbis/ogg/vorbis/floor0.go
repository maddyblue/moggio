package vorbis

import (
	"github.com/jfreymuth/go-vorbis/ogg"
)

type floor0 struct {
	order           uint8
	rate            uint16
	barkMapSize     uint16
	amplitudeBits   uint8
	amplitudeOffset uint8
	bookList        []uint8
}

func (f *floor0) ReadFrom(r *ogg.BitReader) error {
	f.order = r.Read8(8)
	f.rate = r.Read16(16)
	f.barkMapSize = r.Read16(16)
	f.amplitudeBits = r.Read8(6)
	f.amplitudeOffset = r.Read8(8)
	f.bookList = make([]uint8, r.Read8(4)+1)
	for i := range f.bookList {
		f.bookList[i] = r.Read8(8)
	}
	return nil
}

func (f *floor0) Decode(r *ogg.BitReader, books []codebook, n uint32) []float32 {
	amplitude := r.Read32(uint(f.amplitudeBits))
	if amplitude > 0 {
		bookNumber := r.Read8(ilog(len(f.bookList)))
		coefficients := make([]float32, f.order)
		i := 0
	readCoefficients:
		for {
			tempVector := books[f.bookList[bookNumber]].DecodeVector(r)
			for _, c := range tempVector {
				coefficients[i] = c
				i++
				if i >= len(coefficients) {
					break readCoefficients
				}
			}
		}
	}
	//TODO

	return nil
}
