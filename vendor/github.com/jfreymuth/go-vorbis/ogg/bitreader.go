package ogg

type BitReader struct {
	data      []byte
	position  int
	bitOffset uint
	eof       bool
}

func NewBitReader(data []byte) *BitReader {
	return &BitReader{data, 0, 0, false}
}

func NewBitReaderErr(data []byte, err error) (*BitReader, error) {
	return NewBitReader(data), err
}

func (r *BitReader) EOF() bool {
	return r.eof
}

func (r *BitReader) Read1() uint32 {
	if r.position >= len(r.data) {
		r.eof = true
		return 0
	}
	var result uint32
	if r.data[r.position]&(1<<r.bitOffset) != 0 {
		result = 1
	}
	if r.bitOffset < 7 {
		r.bitOffset++
	} else {
		r.bitOffset = 0
		r.position++
	}
	return result
}

func (r *BitReader) Read8(n uint) uint8 {
	if n > 8 {
		panic("ogg: read called with invalid argument")
	}
	var result uint8
	var written uint
	size := n
	for n > 0 {
		if r.position >= len(r.data) {
			r.eof = true
			return 0
		}
		result |= uint8(r.data[r.position]>>r.bitOffset) << written
		written += 8 - r.bitOffset
		if n < 8-r.bitOffset {
			r.bitOffset += n
			break
		}
		n -= 8 - r.bitOffset
		r.bitOffset = 0
		r.position++
	}
	return result &^ (0xFF << size)
}

func (r *BitReader) Read16(n uint) uint16 {
	if n > 16 {
		panic("ogg: read called with invalid argument")
	}
	var result uint16
	var written uint
	size := n
	for n > 0 {
		if r.position >= len(r.data) {
			r.eof = true
			return 0
		}
		result |= uint16(r.data[r.position]>>r.bitOffset) << written
		written += 8 - r.bitOffset
		if n < 8-r.bitOffset {
			r.bitOffset += n
			break
		}
		n -= 8 - r.bitOffset
		r.bitOffset = 0
		r.position++
	}
	return result &^ (0xFFFF << size)
}

func (r *BitReader) Read32(n uint) uint32 {
	if n > 32 {
		panic("ogg: read called with invalid argument")
	}
	var result uint32
	var written uint
	size := n
	for n > 0 {
		if r.position >= len(r.data) {
			r.eof = true
			return 0
		}
		result |= uint32(r.data[r.position]>>r.bitOffset) << written
		written += 8 - r.bitOffset
		if n < 8-r.bitOffset {
			r.bitOffset += n
			break
		}
		n -= 8 - r.bitOffset
		r.bitOffset = 0
		r.position++
	}
	return result &^ (0xFFFFFFFF << size)
}

func (r *BitReader) Read64(n uint) uint64 {
	if n > 64 {
		panic("ogg: read called with invalid argument")
	}
	var result uint64
	var written uint
	size := n
	for n > 0 {
		if r.position >= len(r.data) {
			r.eof = true
			return 0
		}
		result |= uint64(r.data[r.position]>>r.bitOffset) << written
		written += 8 - r.bitOffset
		if n < 8-r.bitOffset {
			r.bitOffset += n
			break
		}
		n -= 8 - r.bitOffset
		r.bitOffset = 0
		r.position++
	}
	return result &^ (0xFFFFFFFFFFFFFFFF << size)
}

func (r *BitReader) ReadBool() bool {
	return r.Read8(1) == 1
}
