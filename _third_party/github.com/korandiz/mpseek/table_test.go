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

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/mjibson/mog/_third_party/github.com/korandiz/mpa"
)

func Test(t *testing.T) {
	var files = []string{
		"test.mp1",
		"test.mp2",
		"test.mp3",
	}

	for _, fn := range files {
		testFile(t, "testdata/"+fn)
	}
}

func testFile(t *testing.T, fn string) {
	seq, err := loadSeq(fn)
	if err != nil {
		t.Errorf("%s: %v", filepath.Base(fn), err)
		return
	}

	rnd, err := loadRnd(fn)
	if err != nil {
		t.Errorf("%s: %v", filepath.Base(fn), err)
		return
	}

	if err := cmp(seq, rnd); err != nil {
		t.Errorf("%s: %v", filepath.Base(fn), err)
		return
	}
}

func loadSeq(fn string) ([2][][]float32, error) {
	var r [2][][]float32

	f, err := os.Open(fn)
	if err != nil {
		return r, err
	}
	defer f.Close()

	d := mpa.Decoder{Input: f}
	for {
		if err := d.DecodeFrame(); err != nil {
			if err == io.EOF {
				err = nil
			}
			return r, err
		}
		for ch := 0; ch < 2; ch++ {
			s := make([]float32, d.NSamples())
			d.ReadSamples(ch, s)
			r[ch] = append(r[ch], s)
		}
	}
}

func loadRnd(fn string) ([2][][]float32, error) {
	var r [2][][]float32

	f, err := os.Open(fn)
	if err != nil {
		return r, err
	}
	defer f.Close()

	t, err := CreateTable(f, 0)
	if err != nil {
		return r, err
	}

	r[0] = make([][]float32, t.NFrames())
	r[1] = make([][]float32, t.NFrames())

	rand.Seed(42)
	for _, i := range rand.Perm(t.NFrames()) {
		x := t.FindSample(int64(i) * int64(t.SamplesPerFrame()))
		if _, err := f.Seek(x.Offset, 0); err != nil {
			return r, err
		}

		d := mpa.Decoder{Input: f}
		for k := 0; k < x.WarmUp; k++ {
			d.DecodeFrame()
		}
		if err := d.DecodeFrame(); err != nil {
			return r, err
		}
		for ch := 0; ch < 2; ch++ {
			s := make([]float32, d.NSamples())
			d.ReadSamples(ch, s)
			r[ch][i] = s
		}
	}

	return r, nil
}

func cmp(a, b [2][][]float32) error {
	for ch := 0; ch < 2; ch++ {
		if len(a[ch]) != len(b[ch]) {
			return fmt.Errorf("frames: %d != %d (ch = %d)",
				len(a[ch]), len(b[ch]), ch)
		}
		for f := range a[ch] {
			if len(a[ch][f]) != len(b[ch][f]) {
				return fmt.Errorf("samples: %d != %d (ch = %d, f = %d)",
					len(a[ch][f]), len(b[ch][f]), ch, f)
			}
			for s := range a[ch][f] {
				if a[ch][f][s] != b[ch][f][s] {
					return fmt.Errorf("mismatch (ch = %d, f = %d, s = %d)",
						ch, f, s)
				}
			}
		}
	}
	return nil
}
