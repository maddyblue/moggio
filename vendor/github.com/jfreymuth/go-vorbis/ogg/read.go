package ogg

import "io"

type Option func(*Reader)

var (
	SkipCRC Option = func(r *Reader) { r.doCRC = false }
)

type Reader struct {
	source      io.Reader
	currentPage page
	index       int
	ready       bool

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

// Length returns the total number of samples that can be decoded.
// The underlying reader must implement io.Seeker.
func (r *Reader) Length() (uint64, error) {
	seeker, _ := r.source.(io.Seeker)
	if seeker == nil {
		return 0, ErrSeek
	}
	position, _ := seeker.Seek(0, 1)
	defer seeker.Seek(position, 0)

	seeker.Seek(0, 0)
	var header pageHeader
	length := uint64(0)
	for {
		err := header.ReadFrom(r.source)
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, err
		}
		length = header.AbsoluteGranulePosition
		seeker.Seek(int64(header.PageSize()), 1)
	}

	return length, nil
}

func (r *Reader) NextPacket() ([]byte, error) {
	if !r.ready {
		err := r.currentPage.ReadFrom(r.source)
		if err != nil {
			return nil, err
		}
		r.index = 0
		r.ready = true
	}
	i := r.index
	if i < r.currentPage.PacketCount() {
		r.index++
		return r.currentPage.Packet(i), nil
	} else {
		r.index = 0
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
