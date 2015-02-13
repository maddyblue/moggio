package server

import (
	"errors"
	"time"
)

type Seek struct {
	b   []float32
	pos int
	sr  time.Duration
	f   func(int) ([]float32, error)
}

func NewSeek(canSeek bool, sr time.Duration, f func(int) ([]float32, error)) *Seek {
	s := Seek{
		f:  f,
		sr: sr,
	}
	if canSeek {
		s.b = make([]float32, 4096)
	}
	return &s
}

func (s *Seek) Read(n int) (b []float32, err error) {
	if s.b == nil {
		b, err = s.f(n)
		s.pos += len(b)
		return
	}
	for len(s.b)-s.pos < n {
		b, err = s.f(n)
		s.b = append(s.b, b...)
		if err != nil {
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
