package vorbis

import (
	"encoding/binary"

	"github.com/jfreymuth/go-vorbis/ogg"
)

const pattern = 0x736962726f76 //"vorbis"

const (
	headerTypeIdentification = 1
	headerTypeComment        = 3
	headerTypeSetup          = 5
)

type identification struct {
	vorbisVersion   uint32
	audioChannels   uint8
	audioSampleRate uint32
	bitrateMaximum  uint32
	bitrateNominal  uint32
	bitrateMinimum  uint32
	blocksize0      uint16
	blocksize1      uint16
}

func (i *identification) ReadFrom(r *ogg.BitReader) error {
	if r.Read8(8) != headerTypeIdentification {
		return ogg.ErrCorruptStream
	}
	if r.Read64(48) != pattern {
		return ogg.ErrCorruptStream
	}
	if r.Read32(32) != 0 {
		return ogg.ErrCorruptStream
	}
	i.audioChannels = r.Read8(8)
	i.audioSampleRate = r.Read32(32)
	i.bitrateMaximum = r.Read32(32)
	i.bitrateNominal = r.Read32(32)
	i.bitrateMinimum = r.Read32(32)
	i.blocksize0 = uint16(1) << r.Read8(4)
	i.blocksize1 = uint16(1) << r.Read8(4)
	if !r.ReadBool() {
		return ogg.ErrCorruptStream
	}
	return nil
}

type comments struct {
	vendor   string
	comments []string
}

func (c *comments) ReadFrom(b []byte) {
	b = b[7:]
	vendorLen := binary.LittleEndian.Uint32(b)
	b = b[4:]
	c.vendor = string(b[:vendorLen])
	b = b[vendorLen:]
	numComments := int(binary.LittleEndian.Uint32(b))
	c.comments = make([]string, numComments)
	b = b[4:]
	for i := 0; i < numComments; i++ {
		commentLen := binary.LittleEndian.Uint32(b)
		b = b[4:]
		c.comments[i] = string(b[:commentLen])
		b = b[commentLen:]
	}
}

type setup struct {
	channels  int
	blocksize [2]int
	codebooks []codebook
	floors    []floor
	residues  []residue
	mappings  []mapping
	modes     []mode

	windows       [2][]float32
	lookup        [2]imdctLookup
	residueBuffer [][]float32
}

type floor interface {
	Decode(*ogg.BitReader, []codebook, uint32) []uint32
	Apply(out []float32, y []uint32)
}

type mapping struct {
	couplingSteps uint16
	angle         []uint8
	magnitude     []uint8
	mux           []uint8
	submaps       []mappingSubmap
}

type mappingSubmap struct {
	floor, residue uint8
}

type mode struct {
	blockflag uint8
	mapping   uint8
}

func (s *setup) ReadFrom(r *ogg.BitReader) error {
	if r.Read8(8) != headerTypeSetup {
		return ogg.ErrCorruptStream
	}
	if r.Read64(48) != pattern {
		return ogg.ErrCorruptStream
	}

	// CODEBOOKS
	s.codebooks = make([]codebook, r.Read16(8)+1)
	for i := range s.codebooks {
		err := s.codebooks[i].ReadFrom(r)
		if err != nil {
			return err
		}
	}

	// TIME DOMAIN TRANSFORMS
	transformCount := r.Read8(6) + 1
	for i := 0; i < int(transformCount); i++ {
		if r.Read16(16) != 0 {
			return ogg.ErrCorruptStream
		}
	}

	// FLOORS
	s.floors = make([]floor, r.Read8(6)+1)
	for i := range s.floors {
		var err error
		switch r.Read16(16) {
		case 0:
			// TODO
			// floor0 is not supported right now, also I haven't found any files that use it.
			/*var f floor0
			err = f.ReadFrom(r)
			s.floors[i] = &f*/
		case 1:
			f := new(floor1)
			err = f.ReadFrom(r)
			s.floors[i] = f
		default:
			err = ogg.ErrCorruptStream
		}
		if err != nil {
			return err
		}
	}

	// RESIDUES
	s.residues = make([]residue, r.Read8(6)+1)
	for i := range s.residues {
		err := s.residues[i].ReadFrom(r)
		if err != nil {
			return err
		}
	}

	// MAPPINGS
	s.mappings = make([]mapping, r.Read8(6)+1)
	for i := range s.mappings {
		m := &s.mappings[i]
		if r.Read16(16) != 0 {
			return ogg.ErrCorruptStream
		}
		if r.ReadBool() {
			m.submaps = make([]mappingSubmap, r.Read8(4)+1)
		} else {
			m.submaps = make([]mappingSubmap, 1)
		}
		if r.ReadBool() {
			m.couplingSteps = r.Read16(8) + 1
			m.magnitude = make([]uint8, m.couplingSteps)
			m.angle = make([]uint8, m.couplingSteps)
			for i := range m.magnitude {
				m.magnitude[i] = r.Read8(ilog(s.channels - 1))
				m.angle[i] = r.Read8(ilog(s.channels - 1))
			}
		}
		if r.Read8(2) != 0 {
			return ogg.ErrCorruptStream
		}
		m.mux = make([]uint8, s.channels)
		if len(m.submaps) > 1 {
			for i := range m.mux {
				m.mux[i] = r.Read8(4)
			}
		}
		for i := range m.submaps {
			r.Read8(8)
			m.submaps[i].floor = r.Read8(8)
			m.submaps[i].residue = r.Read8(8)
		}
	}

	// MODES
	s.modes = make([]mode, r.Read8(6)+1)
	for i := range s.modes {
		m := &s.modes[i]
		m.blockflag = r.Read8(1)
		if r.Read16(16) != 0 {
			return ogg.ErrCorruptStream
		}
		if r.Read16(16) != 0 {
			return ogg.ErrCorruptStream
		}
		m.mapping = r.Read8(8)
	}

	if !r.ReadBool() {
		return ogg.ErrCorruptStream
	}
	return nil
}

func (s *setup) init(i *identification) {
	s.channels = int(i.audioChannels)
	s.blocksize[0] = int(i.blocksize0)
	s.blocksize[1] = int(i.blocksize1)
	s.windows[0] = makeWindow(s.blocksize[0])
	s.windows[1] = makeWindow(s.blocksize[1])
	generateIMDCTLookup(s.blocksize, &s.lookup)
	s.residueBuffer = make([][]float32, s.channels)
	for i := range s.residueBuffer {
		s.residueBuffer[i] = make([]float32, s.blocksize[1]/2)
	}
}
