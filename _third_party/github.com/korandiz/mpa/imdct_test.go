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
	"math"
	"math/rand"
	"testing"
)

func TestImdct(t *testing.T) {
	testImdct(t, new(imdct4Test))
	testImdct(t, new(imdct12Test))
	testImdct(t, new(imdct36Test))
}

func testImdct(t *testing.T, it imdctTester) {
	rand.Seed(42) // make it repeatable
	X := it.input()
	N := 2 * len(X)
	x1 := make([]float64, N)
	max := 0.0
	for i := 0; i < 1000; i++ {
		for k := range X {
			X[k] = 2*rand.Float32() - 1
		}
		directImdct(X, x1)
		x2 := it.transform()
		for k := range x1 {
			max = math.Max(max, math.Abs(x1[k]-float64(x2[k])))
		}
	}

	t.Logf("N = %d, max. difference = %e", N, max)
	if max >= 1.0/(1<<16) {
		t.Fail()
	}
}

func directImdct(in []float32, out []float64) {
	N, Nf := 2*len(in), float64(2*len(in))
	for n := 0; n < N; n++ {
		nf := float64(n)
		out[n] = 0
		for k := 0; k < N/2; k++ {
			kf := float64(k)
			in64 := float64(in[k])
			out[n] += in64 * math.Cos(math.Pi/(2*Nf)*(2*nf+1+Nf/2)*(2*kf+1))
		}
	}
}

type imdctTester interface {
	input() []float32
	transform() []float32
}

type imdct4Test struct {
	in  [2]float32
	out [4]float32
}

func (i *imdct4Test) input() []float32 {
	return i.in[:]
}

func (i *imdct4Test) transform() []float32 {
	imdct4(&i.in, &i.out)
	return i.out[:]
}

type imdct12Test struct {
	in  [6]float32
	out [12]float32
}

func (i *imdct12Test) input() []float32 {
	return i.in[:]
}

func (i *imdct12Test) transform() []float32 {
	imdct12(i.in[:], i.out[:])
	return i.out[:]
}

type imdct36Test struct {
	in  [18]float32
	out [36]float32
}

func (i *imdct36Test) input() []float32 {
	return i.in[:]
}

func (i *imdct36Test) transform() []float32 {
	imdct36(i.in[:], i.out[:])
	return i.out[:]
}
