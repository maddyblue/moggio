package server

import (
	"errors"
	"io"
	"time"
)

type Seek struct {
	b   []byte
	pos int
	sr  time.Duration
	r   io.Reader
}

func NewSeek(canSeek bool, sr time.Duration, r io.Reader) *Seek {
	s := Seek{
		r:  r,
		sr: sr,
	}
	if canSeek {
		s.b = make([]byte, 4096)
	}
	return &s
}

const bytesPerTick = 1

func (s *Seek) Read(n int) (b []byte, err error) {
	b = make([]byte, n)
	if s.b == nil {
		_, err = io.ReadFull(s.r, b)
		s.pos += len(b)
		return
	}
	for len(s.b)-s.pos < n {
		_, err = io.ReadFull(s.r, b)
		s.b = append(s.b, b...)
		if err != nil || len(b) == 0 {
			break
		}
	}
	tot := s.pos + n
	if len(s.b) < tot {
		tot = len(s.b)
	}
	b = s.b[s.pos:tot]
	s.pos = tot
	return
}

var errSeekable = errors.New("cannot seek this file")

// Seek sets the offset for the next Read to offset, relative to the origin
// of the file.
func (s *Seek) Seek(offset time.Duration) error {
	if s.b == nil {
		return errSeekable
	}
	pos := int(offset / s.sr)
	if pos < len(s.b) {
		s.pos = pos
		return nil
	}
	_, err := s.Read(pos - len(s.b))
	if err != nil {
		return err
	}
	s.pos = pos
	return nil
}

func (s *Seek) Pos() time.Duration {
	return s.sr * time.Duration(s.pos)
}
