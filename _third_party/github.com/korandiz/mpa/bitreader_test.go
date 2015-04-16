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

import (
	"bytes"
	"testing"
)

func TestBitreader_readBits(t *testing.T) {
	testData := []byte{
		0x28, 0xc8, 0x28, 0x60, 0x70, 0x40, 0x12, 0x02,
		0x80, 0x2c, 0x01, 0x80, 0x06, 0x80, 0x0e, 0x00,
		0x0f,
	}

	in := []byte(nil)
	for i := 0; i < 4096; i++ {
		in = append(in, testData...)
	}
	r := &bitReader{input: bytes.NewReader(in)}

	for i := 0; i < 1234; i++ {
		d, err := r.readBits(0)
		if err != nil {
			t.Error("Unexpected error:", err)
			return
		}

		if d != 0 {
			t.Errorf("Read 0 bits, got %d. (i = %d)", d, i)
			return
		}
	}

	for i := 0; i < 4096; i++ {
		for j := 0; j < 16; j++ {
			d, err := r.readBits(j + 1)

			if err != nil {
				t.Error("Unexpected error:", err)
				return
			}

			if d != j {
				t.Errorf("i = %d, j = %d, read %d.", i, j, d)
				return
			}
		}
	}

	if _, err := r.readBits(1); err == nil {
		t.Error("Read after EOF returned nil error.")
	}
}

func TestBitreader_readByte(t *testing.T) {
	testData := []byte{
		0xaa, 0x01, 0xaa, 0x02, 0xaa, 0x03, 0xaa, 0x04,
		0xaa, 0x05, 0xaa, 0x06, 0xaa, 0x07, 0xaa, 0x08,
		0xaa, 0xaa, 0x09, 0xaa, 0xaa, 0x0a,
		0xaa, 0xaa, 0x0b, 0xaa, 0xaa, 0x0c,
		0xaa, 0xaa, 0x0d, 0xaa, 0xaa, 0x0e,
		0xaa, 0xaa, 0x0f, 0xaa, 0xaa, 0x10,
		0xaa, 0xaa, 0xaa,
	}

	in := []byte(nil)
	for i := 0; i < 4096; i++ {
		in = append(in, testData...)
	}
	r := &bitReader{input: bytes.NewReader(in)}

	for i := 0; i < 4096; i++ {
		for j := 1; j <= 16; j++ {
			if _, err := r.readBits(j); err != nil {
				t.Error("Unexpected error:", err)
				return
			}

			d, err := r.readByte()
			if err != nil {
				t.Error("Unexpected error:", err)
				return
			}

			if d != j {
				t.Errorf("i = %d, j = %d, read %d.", i, j, d)
				return
			}
		}

		for i := 0; i < 3; i++ {
			d, err := r.readByte()
			if err != nil {
				t.Error("Unexpected error:", err)
				return
			}

			if d != 0xaa {
				t.Errorf("i = %d, read %d.", i, d)
				return
			}
		}
	}

	if _, err := r.readByte(); err == nil {
		t.Error("Read after EOF returned nil error.")
	}
}

func TestBitreader_readBytes(t *testing.T) {
	testData := []byte(nil)
	for i := 0; i < 8; i++ {
		for j := 0; j <= 1440; j++ { // 1440 == max. frame size
			if i > 0 {
				testData = append(testData, 0)
			}
			for k := 0; k < j; k++ {
				testData = append(testData, byte(len(testData)))
			}
		}
	}
	r := &bitReader{input: bytes.NewReader(testData)}

	buff, b := make([]byte, 1440), byte(0)
	for i := 0; i < 8; i++ {
		for j := 0; j <= 1440; j++ {
			if _, err := r.readBits(i); err != nil {
				t.Error("Unexpected error:", err)
				return
			}

			n, err := r.readBytes(buff[0:j])
			if err != nil {
				t.Error("Unexpected error:", err)
				return
			}

			if n != j {
				t.Errorf("Tryed to read %d bytes, read %d.", j, n)
				return
			}

			if i > 0 {
				b++
			}

			for k := range buff[0:n] {
				if buff[k] != b {
					t.Errorf("i = %d, j = %d, k = %d, exptected %d, found %d.",
						i, j, k, b, buff[k])
					return
				}
				b++
			}
		}
	}

	if _, err := r.readBytes(buff[0:1]); err == nil {
		t.Error("Read after EOF returned nil error.")
	}
}

func TestBitreader_syncword3(t *testing.T) {
	testData := []byte{
		0x7f, 0xfd, 0x00,
		0x3f, 0xfe, 0x80,
		0x1f, 0xff, 0x40,
		0x0f, 0xff, 0xa0,
		0x07, 0xff, 0xd0,
		0x03, 0xff, 0xe8,
		0x01, 0xff, 0xf4,
		0x00, 0xff, 0xfa,
		0x00, 0x7f, 0xfd,
	}

	in := []byte(nil)
	for i := 0; i < 4096; i++ {
		in = append(in, testData...)
	}

	r := &bitReader{input: bytes.NewReader(in)}
	for i := 0; i < 4096; i++ {
		for j := range testData {
			sw := r.syncword3()

			if j == 22 && !sw {
				t.Errorf("i = %d, j = %d, syncword3() returned false.", i, j)
				return
			}

			if j != 22 && sw {
				t.Errorf("i = %d, j = %d, syncword3() returned true.", i, j)
				return
			}

			if _, err := r.readByte(); err != nil {
				t.Error("Unexpected error:", err)
				return
			}
		}
	}
}

func TestBitreader_lookahead(t *testing.T) {
	in := make([]byte, 8192)
	for i := 0; i < 100; i++ {
		in[5000+i] = byte(i)
	}
	r := &bitReader{input: bytes.NewReader(in)}

	for i := 0; i < 1000; i++ {
		r.readByte()
	}
	r.readBits(3)

	for i := 0; i < 100; i++ {
		val1 := uint32(i<<24 | (i+1)<<16 | (i+2)<<8 | (i + 3))
		ok1 := 3999+i+3 < len(r.buffer)

		val2, ok2 := r.lookahead(3999 + i)

		if ok1 != ok2 {
			t.Errorf("i = %d, expected %t, got %t", i, ok1, ok2)
			return
		}

		if ok1 && val1 != val2 {
			t.Errorf("i = %d, expected %.8x, got %.8x", i, val1, val2)
			return
		}
	}
}
