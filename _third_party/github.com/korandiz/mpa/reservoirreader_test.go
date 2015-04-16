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

func TestReservoir_setSize_load(t *testing.T) {
	in := []byte(nil)
	for i := 0; i <= 965; i++ {
		for j := i; j < 1930; j++ {
			in = append(in, 1-byte(i%2))
		}
	}
	br := &bitReader{input: bytes.NewReader(in)}
	res := &reservoirReader{stream: br}

	for i := 0; i <= 965; i++ {
		if err := res.setSize(i); err != nil {
			t.Error("setSize failed:", err)
			return
		}
		if err := res.load(1930 - i); err != nil {
			t.Error("load failed:", err)
			return
		}
		for j := 0; j < i; j++ {
			d, err := res.readBits(8)
			if err != nil {
				t.Error("readBits(8) failed:", err)
				return
			}
			if d != i%2 {
				t.Errorf("i = %d, j = %d, expected %d, read %d", i, j, i%2, d)
				return
			}
		}
	}
}

func TestReservoir_loadUntilSyncword(t *testing.T) {
	in := []byte(nil)
	for i := 1; i <= 254; i++ {
		for j := 0; j < i; j++ {
			in = append(in, byte(i))
		}
		in = append(in, 0xff, 0xfa)
	}
	br := &bitReader{input: bytes.NewReader(in)}
	res := &reservoirReader{stream: br}

	for i := 1; i <= 254; i++ {
		if err := res.setSize(0); err != nil {
			t.Error("setSize(0) failed:", err)
			return
		}
		if err := res.loadUntilSyncword(); err != nil {
			t.Error("loadUntilSyncword failed:", err)
			return
		}
		for j := 0; j < i; j++ {
			d, err := res.readBits(8)
			if err != nil {
				t.Error("reservoirReader.readBits(8) failed:", err)
				return
			}
			if d != i {
				t.Errorf("i = %d, j = %d, read %d", i, j, d)
				return
			}
		}
		if _, err := br.readBits(16); err != nil {
			t.Error("bitReader.readBits(16) failed:", err)
			return
		}
	}
}

func TestReservoir_readBits(t *testing.T) {
	testData := []byte{
		0x28, 0xc8, 0x28, 0x60, 0x70, 0x40, 0x12, 0x02,
		0x80, 0x2c, 0x01, 0x80, 0x06, 0x80, 0x0e, 0x00,
		0x0f,
	}

	in := []byte(nil)
	for i := 0; i < 100; i++ {
		in = append(in, testData...)
	}
	br := &bitReader{input: bytes.NewReader(in)}
	res := &reservoirReader{stream: br}
	res.load(len(in))

	for i := 0; i < 1234; i++ {
		d, err := res.readBits(0)
		if err != nil {
			t.Error("Unexpected error:", err)
			return
		}

		if d != 0 {
			t.Errorf("Read 0 bits, got %d. (i = %d)", d, i)
			return
		}
	}

	for i := 0; i < 100; i++ {
		for j := 0; j < 16; j++ {
			d, err := res.readBits(j + 1)

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

	if _, err := res.readBits(1); err == nil {
		t.Error("Read after EOF returned nil error.")
	}
}

func TestReservoir_readCode(t *testing.T) {
	tree := huffmanTree{
		1 * 2, 2 * 2,
		0, 0,
		3 * 2, 4 * 2,
		0, 1,
		5 * 2, 6 * 2,
		0, 2,
		7 * 2, 8 * 2,
		0, 3,
		9 * 2, 10 * 2,
		0, 4,
		11 * 2, 12 * 2,
		0, 5,
		13 * 2, 14 * 2,
		0, 6,
		15 * 2, 16 * 2,
		0, 7,
		17 * 2, 18 * 2,
		0, 8,
		0, 9,
		0, 10,
	}
	testData := []byte{
		0x5b, 0xbd, 0xf7, 0xef, 0xef, 0xf7, 0xfd, 0x6e,
		0xf7, 0xdf, 0xbf, 0xbf, 0xdf, 0xf5, 0xbb, 0xdf,
		0x7e, 0xfe, 0xff, 0x7f, 0xd6, 0xef, 0x7d, 0xfb,
		0xfb, 0xfd, 0xff,
	}

	in := []byte(nil)
	for i := 0; i < 35; i++ {
		in = append(in, testData...)
	}
	br := &bitReader{input: bytes.NewReader(in)}
	res := &reservoirReader{stream: br}
	res.load(len(in))

	for i := 0; i < 4*35; i++ {
		for j := 0; j < 10; j++ {
			d, err := res.readCode(tree)
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

	if _, err := res.readCode(tree); err == nil {
		t.Error("Read after EOF returned nil error.")
	}
}
