// mpa, an MPEG-1 Audio library
// Copyright (C) 2014 KORÁNDI Zoltán <korandi.z@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License, version 3 as
// published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// Please note that, being hungarian, my last name comes before my first
// name. That's why it's in all caps, and not because I like to shout my
// name. So please don't start your emails with "Hi Korandi" or "Dear Mr.
// Zoltan", because it annoys the hell out of me. Thanks.

package mpa

import "io"

// A bitReader wraps an io.Reader and allows its bits to be read one-by-one, in
// MSB-first order.
// The buffer must be large enough to hold two frames' worth of data, or
// Decoder.trueHeader won't work properly.
type bitReader struct {
	input   io.Reader
	buffer  [4096]byte
	head    int  // index of the first unread byte in the buffer
	limit   int  // index of the first byte not in the buffer (head <= limit)
	current byte // byte being read, left-shifted by the # of bits read so far
	bits    int  // number of unread bits in 'current' (0 <= bits < 8)
}

// refill loads the next chunk from the input to the buffer, and updates 'head'
// and 'limit' accordingly. Any unread data in the buffer is moved to the
// beginning.
func (rd *bitReader) refill() error {
	copy(rd.buffer[0:], rd.buffer[rd.head:rd.limit])
	rd.limit -= rd.head
	rd.head = 0

	if rd.limit == len(rd.buffer) {
		return nil
	}

	for k := 0; k < 8; k++ {
		if n, err := rd.input.Read(rd.buffer[rd.limit:]); n > 0 {
			// If any data has been successfully read, we process it, regardless
			// of whether or not err is nil. Non-nil errors are silently
			// ignored, but the next call to Read will likely fail and return no
			// data, causing refill to fail as well. Or the error may just go
			// away, which is okay, too.
			rd.limit += n
			if rd.limit > len(rd.buffer) {
				rd.limit = len(rd.buffer)
			}
			return nil
		} else if err != nil {
			return err
		}
	}
	return io.ErrNoProgress
}

// readBits reads the next n bits from the stream and returns them as an
// integer.
func (rd *bitReader) readBits(n int) (int, error) {
	if n <= 0 {
		return 0, nil
	}

	retval, bits := 0, rd.bits
	for {
		if bits == 0 {
			if rd.head == rd.limit {
				if err := rd.refill(); err != nil {
					return 0, err
				}
			}
			bits, rd.current = 8, rd.buffer[rd.head]
			rd.head++
		}

		if n <= bits {
			retval <<= uint(n)
			retval |= int(rd.current >> uint(8-n))
			rd.current <<= uint(n)
			bits -= n
			rd.bits = bits
			return retval, nil
		} else {
			retval <<= uint(bits)
			retval |= int(rd.current >> uint(8-bits))
			n -= bits
			bits = 0
		}
	}
}

// readByte skips to the next byte boundary, reads the next byte, and returns it
// as an integer.
func (rd *bitReader) readByte() (int, error) {
	if rd.head == rd.limit {
		if err := rd.refill(); err != nil {
			return 0, err
		}
	}

	rd.bits = 0
	rd.current = rd.buffer[rd.head]
	rd.head++
	return int(rd.current), nil
}

// readBytes skips to the next byte boundary, reads up to len(dst) bytes to
// dst, and returns the number of bytes read.
func (rd *bitReader) readBytes(dst []byte) (int, error) {
	rd.bits = 0
	total := 0

	for len(dst) > 0 {
		if rd.head == rd.limit {
			if err := rd.refill(); err != nil {
				return total, err
			}
		}

		n := copy(dst, rd.buffer[rd.head:rd.limit])
		dst = dst[n:]
		rd.head += n
		total += n
	}

	return total, nil
}

// syncword3 returns true if and only if a Layer III syncword (the bit string
// 1111 1111 1111 101) starts at the next byte boundary.
func (rd *bitReader) syncword3() bool {
	if rd.head+1 >= rd.limit {
		rd.refill()

		// Any error reported by refill is ignored, otherwise a perfectly
		// legitimate EOF could prevent the last frame from being decoded. In
		// other words, there's nothing to look ahead at beyond the EOF, and
		// that's not an error.

		if rd.head+1 >= rd.limit {
			return false
		}
	}

	return rd.buffer[rd.head] == 0xff && rd.buffer[rd.head+1]&0xfe == 0xfa
}

// lookahead returns the 32-bit word starting n bytes ahead of the current read
// position.
func (rd *bitReader) lookahead(n int) (uint32, bool) {
	if rd.head+n+4 > rd.limit {
		rd.refill()
		if rd.head+n+4 > rd.limit {
			return 0, false
		}
	}

	var r uint32
	for i := 0; i < 4; i++ {
		r = r<<8 | uint32(rd.buffer[rd.head+n+i])
	}
	return r, true
}
