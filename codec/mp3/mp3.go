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
	return &MP3{
		r: bufio.NewReader(r),
	}, nil
}

func (m *MP3) Scan() bool {
	pos := 0
	var f Frame
	for {
		b, err := m.r.ReadByte()
		if err != nil {
			m.err = err
			return false
		}
		switch pos {
		case 0:
			if b == 0xff {
				pos++
			}
		case 1:
			if b&0xe0 == 0 {
				pos = 0
				break
			}
			pos++
			f = Frame{
				Version: Version(b & 0x18 >> 3),
				Layer:   Layer(b & 0x6 >> 1),
				CRC:     b&0x1 != 0,
			}
		case 2:
			pos++
			f.Bitrate = Bitrate(b & 0xf0 >> 4)
			f.Sampling = Sampling(b & 0xc >> 2)
			f.Padding = b&0x2 != 0
		case 3:
			f.Mode = Mode(b & 0xc >> 4)
			f.Emphasis = Emphasis(b & 0x3)
			if !f.Valid() {
				pos = 0
				break
			}
			m.frame = &f
			return true
		default:
			pos = 0
		}
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

type Layer byte

const (
	LayerI   Layer = 3
	LayerII        = 2
	LayerIII       = 1
)

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
