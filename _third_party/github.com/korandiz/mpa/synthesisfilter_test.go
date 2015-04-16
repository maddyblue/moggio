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
// name. So please don't start your emails with "Hi Korandi" or "Dear
// Mr. Zoltan", because it annoys the hell out of me. Thanks.

package mpa

import (
	"math"
	"math/rand"
	"testing"
)

func TestSynthesisFilter(t *testing.T) {
	rand.Seed(42) // make it repeatable
	var (
		f1   directSynthesisFilter
		in1  [32]float64
		out1 [32]float64
		f2   synthesisFilter
		x2   [32]float32
		max  float64
	)
	for i := 0; i < 2048; i++ {
		for j := 0; j < 32; j++ {
			in1[j] = 2*rand.Float64() - 1
			x2[j] = float32(in1[j])
		}
		f1.filter(&in1, &out1)
		f2.filter(x2[:])
		for j := 0; j < 32; j++ {
			max = math.Max(max, math.Abs(out1[j]-float64(x2[j])))
		}
	}

	t.Logf("max. difference = %e", max)
	if max >= 1.0/(1<<16) {
		t.Fail()
	}
}

type directSynthesisFilter [1024]float64

func (v *directSynthesisFilter) filter(in, out *[32]float64) {
	for i := 1023; i >= 64; i-- {
		v[i] = v[i-64]
	}

	for i := 0; i <= 63; i++ {
		v[i] = 0
		iFl := float64(i)
		for k := 0; k <= 31; k++ {
			kFl := float64(k)
			v[i] += math.Cos((16+iFl)*(2*kFl+1)*math.Pi/64.0) * float64(in[k])
		}
	}

	var u [512]float64
	for i := 0; i <= 7; i++ {
		for j := 0; j <= 31; j++ {
			u[i*64+j] = v[i*128+j]
			u[i*64+32+j] = v[i*128+96+j]
		}
	}

	var w [512]float64
	for i := 0; i <= 511; i++ {
		w[i] = u[i] * float64(synthesisWindow[i])
	}

	for j := 0; j <= 31; j++ {
		out[j] = 0
		for i := 0; i <= 15; i++ {
			out[j] += w[j+32*i]
		}
	}
}
