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

func TestDct(t *testing.T) {
	testDct(t, new(dct2Test))
	testDct(t, new(dct4Test))
	testDct(t, new(dct8Test))
	testDct(t, new(dct16Test))
	testDct(t, new(dct32Test))
}

func testDct(t *testing.T, dt dctTester) {
	rand.Seed(42) // make it repeatable
	max, x := 0.0, dt.slice()
	for i := 0; i < 1000; i++ {
		for i := range x {
			x[i] = 2*rand.Float32() - 1
		}
		X := directDct(x)
		dt.transform()
		for i := range x {
			max = math.Max(max, math.Abs(X[i]-float64(x[i])))
		}
	}

	t.Logf("N = %d, max. difference = %e", len(x), max)
	if max >= 1.0/(1<<16) {
		t.Fail()
	}
}

func directDct(x []float32) []float64 {
	N, NFl := len(x), float64(len(x))
	X := make([]float64, N)
	for n := 0; n < N; n++ {
		nFl := float64(n)
		for k := 0; k < N; k++ {
			kFl := float64(k)
			X[n] += float64(x[k]) * math.Cos(math.Pi*(2*kFl+1)*nFl/(2*NFl))
		}
	}
	return X
}

type dctTester interface {
	slice() []float32
	transform()
}

type dct2Test [2]float32

func (d *dct2Test) slice() []float32 {
	return d[:]
}
func (d *dct2Test) transform() {
	dct2((*[2]float32)(d))
}

type dct4Test [4]float32

func (d *dct4Test) slice() []float32 {
	return d[:]
}
func (d *dct4Test) transform() {
	dct4((*[4]float32)(d))
}

type dct8Test [8]float32

func (d *dct8Test) slice() []float32 {
	return d[:]
}
func (d *dct8Test) transform() {
	dct8((*[8]float32)(d))
}

type dct16Test [16]float32

func (d *dct16Test) slice() []float32 {
	return d[:]
}
func (d *dct16Test) transform() {
	dct16((*[16]float32)(d))
}

type dct32Test [32]float32

func (d *dct32Test) slice() []float32 {
	return d[:]
}
func (d *dct32Test) transform() {
	dct32(d[:])
}
