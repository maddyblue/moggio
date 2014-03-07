package mp3

import (
	"bufio"
	"io"
)

type MP3 struct {
	r     *bufio.Reader
	frame *Frame
	err   error
}

func New(r io.Reader) (*MP3, error) {
	m := &MP3{
		r: bufio.NewReader(r),
	}
	b, err := m.r.Peek(10)
	if err != nil {
		return nil, err
	}
	if b[0] == 'I' && b[1] == 'D' && b[2] == '3' && b[3] < 0xff && b[4] < 0xff && b[6] < 0x80 && b[7] < 0x80 && b[8] < 0x80 && b[9] < 0x80 {
		var sz uint32
		sz = uint32(b[9] & 0x7f)
		sz += uint32(b[8]&0x7f) << 7
		sz += uint32(b[7]&0x7f) << 14
		sz += uint32(b[6]&0x7f) << 21
		sz += uint32(len(b))
		for ; sz > 0; sz-- {
			if _, err := m.r.ReadByte(); err != nil {
				return nil, err
			}
		}
	}
	return m, nil
}

func (m *MP3) Scan() bool {
	var f Frame
	for {
		b, err := m.r.Peek(4)
		if err != nil {
			m.err = err
			return false
		}
		switch b[0] {
		case 0xff:
			if b[1]&0xe0 != 0xe0 {
				break
			}
			f = Frame{
				Version:  Version(b[1] & 0x18 >> 3),
				Layer:    Layer(b[1] & 0x6 >> 1),
				CRC:      b[1]&0x1 != 0,
				Bitrate:  Bitrate(b[2] & 0xf0 >> 4),
				Sampling: Sampling(b[2] & 0xc >> 2),
				Padding:  b[2]&0x2 != 0,
				Mode:     Mode(b[3] & 0xc >> 4),
				Emphasis: Emphasis(b[3] & 0x3),
			}
			if !f.Valid() {
				break
			}
			m.frame = &f
			m.r.Read(b)
			return true
		}
		m.r.ReadByte()
	}
}

func (m *MP3) Err() error {
	if m.err == io.EOF {
		return nil
	}
	return m.err
}

func (m *MP3) Frame() *Frame {
	return m.frame
}

type Frame struct {
	Version
	Layer
	CRC bool
	Bitrate
	Sampling
	Padding bool
	Mode
	Emphasis
}

func (f *Frame) Valid() bool {
	if f.Version != MPEG1 {
		return false
	}
	if f.Layer != LayerI {
		return false
	}
	if f.Bitrate == 0xff || f.Bitrate == 0 {
		return false
	}
	return true
}

type Version byte

const (
	MPEG1 Version = 3
	MPEG2         = 2
)

func (v Version) String() string {
	switch v {
	case MPEG1:
		return "MPEG1"
	case MPEG2:
		return "MPEG2"
	default:
		return "unknown"
	}
}

type Layer byte

const (
	LayerI   Layer = 3
	LayerII        = 2
	LayerIII       = 1
)

func (l Layer) String() string {
	switch l {
	case LayerI:
		return "layer I"
	case LayerII:
		return "layer II"
	case LayerIII:
		return "layer III"
	default:
		return "unknown"
	}
}

type Bitrate byte

type Sampling byte

type Mode byte

const (
	ModeStereo Mode = 0
	ModeJoint       = 1
	ModeDual        = 2
	ModeSingle      = 3
)

type Emphasis byte

const (
	EmphasisNone  Emphasis = 0
	Emphasis50_15          = 1
	EmphasisCCIT           = 3
)
