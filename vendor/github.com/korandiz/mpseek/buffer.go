// mpseek, a library to support seeking MPEG Audio files
// Copyright (C) 2015 KORÁNDI Zoltán <korandi.z@gmail.com>
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

package mpseek

import "io"

// A buffer is a temporary storage which provides some high-level methods for
// dealing with the stream. The capacity, len(data) must be at least twice the
// maximum possible frame size. offs is the number of bytes already consumed.
// data[head:limit] contains valid, unconsumed data.
type buffer struct {
	input io.Reader
	data  [4096]byte
	head  int
	limit int
	offs  int64
	err   error
}

// load makes sure there are at least n bytes in the buffer. n must not exceed
// the capacity.
func (b *buffer) load(n int) error {
	for k := 0; b.limit-b.head < n && b.err == nil; {
		copy(b.data[0:], b.data[b.head:b.limit])
		b.limit -= b.head
		b.head = 0

		m, err := b.input.Read(b.data[b.limit:])
		if m > 0 {
			b.limit += m
			if b.limit > len(b.data) {
				b.limit = len(b.data)
			}
			k = 0
		} else if err == nil {
			if k++; k == 8 {
				err = io.ErrNoProgress
			}
		}
		b.err = err
	}

	if b.limit-b.head >= n {
		return nil
	}
	return b.err
}

// lookahead returns the potential frame header n bytes ahead of the current
// read position without advancing the stream cursor. n+4 must not exceed the
// buffer's capacity.
func (b *buffer) lookahead(n int) (h header, ok bool) {
	if err := b.load(n + 4); err != nil {
		return 0, false
	}

	for i := 0; i < 4; i++ {
		h = h<<8 | header(b.data[b.head+n+i])
	}
	return h, true
}

// skip discards the next n bytes of the stream. n must not exceed the buffer's
// capacity.
func (b *buffer) skip(n int) error {
	if err := b.load(n); err != nil {
		b.offs += int64(b.limit - b.head)
		b.head = b.limit
		return err
	}

	b.offs += int64(n)
	b.head += n
	return nil
}

// synchronize reads the stream until a frame header is found. False syncwords
// are ignored.
func (b *buffer) synchronize() (header, error) {
	var h header

retry:
	h &^= 0xff << 24

	for h.syncword() != 0xfff {
		if err := b.load(1); err != nil {
			return 0, err
		}
		h = h<<8 | header(b.data[b.head])
		b.head++
		b.offs++
	}

	if !h.valid() {
		goto retry
	}
	if h2, ok := b.lookahead(h.frameSize() - 4); ok {
		if !h2.valid() {
			goto retry
		}
		if h3, ok := b.lookahead(h.frameSize() - 4 + h2.frameSize()); ok {
			if !h3.valid() {
				goto retry
			}
		}
	}

	return h, nil
}

// peekMDB returns the main_data_begin field of the frame whose header h has
// just been read out of the stream. The stream cursor is not advanced.
func (b *buffer) peekMDB(h header) (int, error) {
	n := 0
	if h.protectionBit() == 0 {
		n = 2
	}
	if err := b.load(n + 2); err != nil {
		return 0, err
	}
	return int(b.data[b.head+n])<<1 | int(b.data[b.head+n+1])>>7, nil
}
