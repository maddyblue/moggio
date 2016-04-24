package ogg

import "io"

type Option func(*Reader)

var (
	SkipCRC Option = func(r *Reader) { r.doCRC = false }
)

type Reader struct {
	source         io.Reader
	currentPage    page
	index          int
	pageFirstIndex int
	err            error
	ready          bool

	doCRC bool
}

func NewReader(in io.Reader, options ...Option) *Reader {
	r := &Reader{
		source: in,
		doCRC:  true,
	}
	for _, option := range options {
		option(r)
	}
	return r
}

func (r *Reader) NextPacket() ([]byte, error) {
	if !r.ready {
		err := r.currentPage.ReadFrom(r.source)
		if err != nil {
			return nil, err
		}
		r.ready = true
	}
	i := r.index - r.pageFirstIndex
	if i < r.currentPage.PacketCount() {
		r.index++
		return r.currentPage.Packet(i), nil
	} else {
		r.pageFirstIndex = r.index
		rest := r.currentPage.Rest()
		err := r.currentPage.ReadFrom(r.source)
		if err != nil {
			return nil, err
		}
		if r.doCRC {
			crc := crcUpdate(r.currentPage.header.headerChecksum, r.currentPage.content)
			if crc != r.currentPage.header.PageChecksum {
				return nil, ErrChecksum
			}
		}
		err = r.currentPage.Prepend(rest)
		if err != nil {
			return nil, err
		}
		// the recursion is needed to handle packets spanning multiple page boundaries correctly
		return r.NextPacket()
	}
}
