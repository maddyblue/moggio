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
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"testing"
)

var PS = string(os.PathSeparator)

func TestDecoder_Compliance1(t *testing.T) {
	testCompliance12(t, "fl1", 2, 18816, 4)
	testCompliance12(t, "fl2", 2, 18816, 4)
	testCompliance12(t, "fl3", 2, 18816, 4)
	testCompliance12(t, "fl4", 1, 18816, 4)
	testCompliance12(t, "fl5", 2, 18816, 4)
	testCompliance12(t, "fl6", 2, 18816, 4)
	testCompliance12(t, "fl7", 2, 24192, 4)
	testCompliance12(t, "fl8", 2, 18816, 4)
}

func TestDecoder_Compliance2(t *testing.T) {
	testCompliance12(t, "fl10", 2, 56448, 1)
	testCompliance12(t, "fl11", 2, 56448, 1)
	testCompliance12(t, "fl12", 2, 56448, 1)
	testCompliance12(t, "fl13", 1, 56448, 1)
	testCompliance12(t, "fl14", 2, 18432, 1)
	testCompliance12(t, "fl15", 2, 18432, 1)
	testCompliance12(t, "fl16", 2, 72576, 1)
}

func TestDecoder_Compliance3(t *testing.T) {
	testCompliance3(t, "compl", 1, 248832)
	testCompliance3(t, "hecommon", 2, 33408)
	testCompliance3(t, "he_32khz", 1, 171648)
	testCompliance3(t, "he_44khz", 1, 471168)
	testCompliance3(t, "he_48khz", 1, 171648)
	testCompliance3(t, "he_free", 2, 77184)
	testCompliance3_heMode(t)
	testCompliance3(t, "si", 1, 134784)
	testCompliance3(t, "si_block", 1, 72576)
	testCompliance3(t, "si_huff", 1, 85248)
	testCompliance3(t, "sin1k0db", 2, 362880)
}

func testCompliance12(t *testing.T, fn string, nch int, n, k int) {
	t.Log("==", fn, "==")

	in, err := loadHexFile(fn+".mpg", k)
	if err != nil {
		t.Log("    ", err)
		return
	}

	ref, err := loadPcmFile(fn+".pcm", nch)
	if err != nil {
		t.Log("    ", err)
		return
	}

	test, err := decodeTestFile(in, nch)
	if err != nil {
		t.Error("    ", err)
		return
	}

	testRms(t, ref, test, nch, n, 1)
}

func testCompliance3(t *testing.T, fn string, nch, n int) {
	t.Log("==", fn, "==")

	in, err := ioutil.ReadFile("iso11172-4" + PS + fn + ".bit")
	if err != nil {
		t.Log("    ", err)
		return
	}

	ref, err := loadPcmFile(fn+".hex", nch)
	if err != nil {
		t.Log("    ", err)
		return
	}

	test, err := decodeTestFile(in, nch)
	if err != nil {
		t.Error("    ", err)
		return
	}

	testRms(t, ref, test, nch, n, 1)
}

func testCompliance3_heMode(t *testing.T) {
	// he_mode requires special treatment, because it contains both stereo and
	// mono frames.

	t.Log("== he_mode ==")

	in, err := ioutil.ReadFile("iso11172-4" + PS + "he_mode.bit")
	if err != nil {
		t.Log("    ", err)
		return
	}

	ref0, err := loadPcmFile("he_mode.hex", 1)
	if err != nil {
		t.Log("    ", err)
		return
	}

	// Frames 1-10 and 111-127 are in mono. Fix that.
	n := 146304
	ref := [2][]int32{make([]int32, n), make([]int32, n)}
	for frame, s, d := 1, 0, 0; frame <= 127; frame++ {
		if frame <= 10 || frame >= 111 && frame <= 127 {
			for t := 0; t < 1152; t++ {
				ref[0][d] = ref0[0][s]
				ref[1][d] = ref0[0][s]
				d++
				s++
			}
		} else {
			for t := 0; t < 1152; t++ {
				ref[0][d] = ref0[0][s]
				ref[1][d] = ref0[0][s+1]
				d++
				s += 2
			}
		}
	}

	test, err := decodeTestFile(in, 2)
	if err != nil {
		t.Error("    ", err)
		return
	}

	// The reference output for Layer III test streams (except for compl.hex)
	// was generated with l3dec which seems to do no alias reduction on the
	// lower subbands of granules with mixed blocks. To prevent the test from
	// failing, we set an error margin 6x greater than normal.
	testRms(t, ref, test, 2, n, 6)
}

func loadHexFile(fn string, k int) ([]byte, error) {
	file, err := os.Open("iso11172-4" + PS + fn)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data := []byte(nil)
	for {
		var x int
		n, err := fmt.Fscanf(file, "%X\n", &x)
		if n != 0 {
			if k == 4 {
				data = append(data, byte((x>>24)&0xff))
				data = append(data, byte((x>>16)&0xff))
				data = append(data, byte((x>>8)&0xff))
			}
			data = append(data, byte(x&0xff))
		}
		if err == io.EOF {
			return data, nil
		} else if err != nil {
			return nil, err
		}
	}
}

func loadPcmFile(fn string, nch int) ([2][]int32, error) {
	data := [2][]int32{}
	file, err := os.Open("iso11172-4" + PS + fn)
	if err != nil {
		return data, err
	}
	defer file.Close()
	for {
		for ch := 0; ch < nch; ch++ {
			var x int32
			n, err := fmt.Fscanf(file, "%X\n", &x)
			if n != 0 {
				data[ch] = append(data[ch], x)
			}
			if err == io.EOF {
				return data, nil
			} else if err != nil {
				return data, err
			}
		}
	}

}

func decodeTestFile(in []byte, nch int) ([2][]int32, error) {
	d := &Decoder{Input: bytes.NewReader(in)}
	tmp := make([]float32, 1152)
	data := [2][]int32{}
	for {
		if err := d.DecodeFrame(); err != nil {
			if _, ms := err.(MalformedStream); ms {
				continue
			} else if err == io.EOF || err == io.ErrUnexpectedEOF {
				return data, nil
			} else {
				return data, err
			}
		}
		N := d.NSamples()
		for ch := 0; ch < nch; ch++ {
			d.ReadSamples(ch, tmp)
			for t := 0; t < N; t++ {
				x := int32((tmp[t]+1)*(1<<23)) ^ (1 << 23)
				if x == 1<<24 {
					x = (1 << 24) - 1
				}
				data[ch] = append(data[ch], x)
			}
		}
	}
}

func testRms(t *testing.T, ref, test [2][]int32, nch, n int, k float64) {
	maxRms := 1 / (math.Pow(2, 15) * math.Sqrt(12))
	maxMax := math.Pow(2, -14)

	if len(test[0]) < n {
		t.Errorf("    # of samples/channel: %d, expected: %d", len(test[0]), n)
		t.Log("  [FAIL]")
		return
	}

	fail := false
	for ch := 0; ch < nch; ch++ {
		rms, max := 0.0, 0.0
		for t := 0; t < n; t++ {
			s1 := float64(ref[ch][t]^(1<<23))/(1<<23) - 1
			s2 := float64(test[ch][t]^(1<<23))/(1<<23) - 1
			d := s1 - s2
			if d < 0 {
				d *= -1
			}
			if d > max {
				max = d
			}
			rms += d * d
		}
		rms = math.Sqrt(rms / float64(n))
		rms /= maxRms
		max /= maxMax
		t.Logf("    Channel %d: rms=%f, max=%f", ch, rms, max)
		if rms >= k || max > k {
			t.Fail()
			fail = true
		}
	}
	if fail {
		t.Error("  [FAIL]")
	}
}

func TestDecoder_errorTolerance(t *testing.T) {
	// This test never explicitly fails. If it doesn't panic, it's okay.

	var files = []string{
		// Layer I
		"fl1.mpg",
		"fl2.mpg",
		"fl3.mpg",
		"fl4.mpg",
		"fl5.mpg",
		"fl6.mpg",
		"fl7.mpg",
		"fl8.mpg",
		"fl10.mpg",
		"fl11.mpg",
		"fl12.mpg",
		"fl13.mpg",
		"fl14.mpg",
		"fl15.mpg",
		"fl16.mpg",
		"compl.bit",
		"he_32khz.bit",
		"he_44khz.bit",
		"he_48khz.bit",
		"he_free.bit",
		"he_mode.bit",
		"hecommon.bit",
		"si.bit",
		"si_block.bit",
		"si_huff.bit",
		"sin1k0db.bit",
	}
	n := 100
	if testing.Short() {
		n = 1
	}
	for _, fn := range files {
		fuzz(t, fn, n, 1e-4)
	}
}

func fuzz(t *testing.T, fn string, n int, p float64) {
	rand.Seed(42)
	in1, err := ioutil.ReadFile("iso11172-4" + PS + fn)
	if err != nil {
		t.Logf("Could not read file %s: %s", fn, err)
		return
	}
	in := []byte(nil)
	for i := 0; i < n; i++ {
		in = append(in, in1...)
	}
	d := Decoder{Input: &fuzzer{bytes.NewReader(in), p}}
	tmp := make([]float32, 2000)

	for {
		err := d.DecodeFrame()
		if _, ms := err.(MalformedStream); err != nil && !ms {
			break
		}
		d.Bitrate()
		d.Copyrighted()
		d.Emphasis()
		d.Layer()
		d.Mode()
		d.NChannels()
		d.NSamples()
		d.Original()
		d.SamplingFrequency()
		d.ReadSamples(ChLeft, tmp)
		d.ReadSamples(ChRight, tmp)
	}
}

type fuzzer struct {
	r io.Reader
	p float64
}

func (f *fuzzer) Read(p []byte) (int, error) {
	n, err := f.r.Read(p)
	for i := 0; i < n; i++ {
		if rand.Float64() < 8*f.p {
			p[i] ^= 1 << uint(rand.Intn(8))
		}
	}
	return n, err
}

func TestDecoder_junkTolerance(t *testing.T) {
	// This test never explicitly fails. If it doesn't panic, it's okay.

	N := 1 << 30
	if testing.Short() {
		N = 1 << 23
	}
	rand.Seed(42)
	d := &Decoder{Input: &junkReader{N}}
	tmp := make([]float32, 2000)
	for n := 0; ; n++ {
		err := d.DecodeFrame()
		if _, ms := err.(MalformedStream); err != nil && !ms {
			t.Logf("Decoder.DecodeFrame called %d times.", n)
			break
		}
		d.Bitrate()
		d.Copyrighted()
		d.Emphasis()
		d.Layer()
		d.Mode()
		d.NChannels()
		d.NSamples()
		d.Original()
		d.SamplingFrequency()
		d.ReadSamples(ChLeft, tmp)
		d.ReadSamples(ChRight, tmp)
	}
}

type junkReader struct {
	n int
}

func (r *junkReader) Read(p []byte) (int, error) {
	n := 0
	for n < len(p) && n < r.n {
		p[n] = byte(rand.Int())
		n++
	}
	r.n -= n
	if n < len(p) {
		return n, io.EOF
	}

	return n, nil
}
