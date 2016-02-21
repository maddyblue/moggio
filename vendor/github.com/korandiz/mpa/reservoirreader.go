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

// A reservoirReader is an intermediate buffer for the main data stream in Layer
// III. Before the decoder starts reading the main data for a frame, it makes
// sure all main data between the byte pointed by main_data_begin and the last
// byte of the frame being decoded is in the reservoir. Thus, the required
// buffer size is the maximum frame size, minus the header size, minus the
// minimum length of the side information, plus the maximum value of
// main_data_begin.
type reservoirReader struct {
	stream  *bitReader
	buffer  [1441 - 4 - 17 + 511]byte
	head    int  // index of the first unread byte in the buffer
	limit   int  // index of the first byte not in the buffer
	current byte // byte being read, left-shifted by the # of bits read so far
	bits    int  // number of unread bits in 'current' (0 <= bits < 8)
}

// setSize moves the last up to n bytes in the reservoir to the beginning of the
// buffer and discards all preceding data. It also sets the current read
// position to the beginning of the buffer, which may result in the main data
// stream essentially being rewound.
func (rd *reservoirReader) setSize(n int) error {
	rd.bits = 0
	rd.head = 0
	if n > rd.limit {
		return MalformedStream("not enough main data")
	}
	copy(rd.buffer[0:], rd.buffer[rd.limit-n:rd.limit])
	rd.limit = n
	return nil
}

// load moves the next n bytes from the input stream to the reservoir.
func (rd *reservoirReader) load(n int) error {
	m, err := rd.stream.readBytes(rd.buffer[rd.limit : rd.limit+n])
	rd.limit += m
	return err
}

// loadUntilSyncword moves bytes from the input stream to the reservoir
// one-by-one until a syncword is found.
func (rd *reservoirReader) loadUntilSyncword() error {
	var err error
	var x int
	overflow := false

	for !rd.stream.syncword3() {
		if x, err = rd.stream.readByte(); err != nil {
			break
		}
		if rd.limit == len(rd.buffer) {
			// Even if the data overflows the buffer, we keep reading until a
			// syncword is found. First, the decoder stays in sync this way.
			// Second, this ensures that we have the most recent data in the
			// reservoir, so chances are higher it will be possible to decode
			// the next frame.
			rd.limit, overflow = 0, true
		}
		rd.buffer[rd.limit] = byte(x)
		rd.limit++
	}

	if overflow {
		tmp := rd.buffer
		n := copy(rd.buffer[0:], tmp[rd.limit:])
		copy(rd.buffer[n:], tmp[:rd.limit])
		rd.head, rd.limit = 0, len(rd.buffer)
		if err == nil {
			err = MalformedStream("reservoir overflow")
		}
	}

	return err
}

// readBits reads the next n bits from the reservoir and returns them as an
// integer.
func (rd *reservoirReader) readBits(n int) (int, error) {
	if n <= 0 {
		return 0, nil
	}

	retval, bits := 0, rd.bits
	for {
		if bits == 0 {
			if rd.head == rd.limit {
				return 0, MalformedStream("reservoir overread")
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

// readCode reads one Huffman coded value from the reservoir using the specified
// code tree.
func (rd *reservoirReader) readCode(tree huffmanTree) (int, error) {
	n, bits, current := uint32(0), rd.bits, rd.current
	for tree[n] != 0 {
		if bits == 0 {
			if rd.head == rd.limit {
				rd.bits = 0
				return 0, MalformedStream("not enough Huffman data")
			}
			bits, current = 8, rd.buffer[rd.head]
			rd.head++
		}
		n = tree[n+uint32(current>>7)]
		current <<= 1
		bits--
	}
	rd.bits, rd.current = bits, current
	return int(tree[n+1]), nil
}
