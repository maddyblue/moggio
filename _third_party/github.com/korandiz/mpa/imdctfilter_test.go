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

func TestImdctFilter(t *testing.T) {
	for typ := 0; typ <= 3; typ++ {
		in1, out1 := make([]float32, 18), make([]float32, 18)
		in2, out2 := make([]float64, 18), make([]float64, 18)
		f1, f2 := imdctFilter{}, directImdctFilter{}
		rand.Seed(42)
		max := 0.0
		for i := 0; i < 1000; i++ {
			for j := 0; j < 18; j++ {
				x := 2*rand.Float64() - 1
				in1[j] = float32(x)
				in2[j] = x
			}
			f1.filter(in1, out1, typ)
			f2.filter(in2, out2, typ)
			for j := 0; j < 18; j++ {
				max = math.Max(max, math.Abs(float64(out1[j])-out2[j]))
			}
		}
		t.Logf("Type = %d, max. difference = %e", typ, max)
		if max >= 1.0/(1<<16) {
			t.Fail()
		}
	}
}

type directImdctFilter struct {
	prev, curr [36]float64
}

func (f *directImdctFilter) filter(X []float64, x []float64, typ int) {
	if typ != 2 {
		tmp := make([]float64, 36)
		for i := 0; i < 36; i++ {
			iF := float64(i)
			for k := 0; k < 18; k++ {
				kF := float64(k)
				tmp[i] += X[k] * math.Cos(math.Pi/72*(2*iF+19)*(2*kF+1))
			}
			var w float64
			switch typ {
			case 0:
				w = math.Sin(math.Pi / 36 * (iF + 0.5))
			case 1:
				switch {
				case i <= 17:
					w = math.Sin(math.Pi / 36 * (iF + 0.5))
				case i <= 23:
					w = 1
				case i <= 29:
					w = math.Sin(math.Pi / 12 * (iF - 17.5))
				default:
					w = 0
				}
			case 3:
				switch {
				case i <= 5:
					w = 0
				case i <= 11:
					w = math.Sin(math.Pi / 12 * (iF - 5.5))
				case i <= 17:
					w = 1
				default:
					w = math.Sin(math.Pi / 36 * (iF + 0.5))
				}
			}
			f.curr[i] = w * tmp[i]
		}
	} else {
		var (
			Y [3][]float64
			y [3][]float64
		)
		for w := 0; w < 3; w++ {
			Y[w] = make([]float64, 18)
			y[w] = make([]float64, 36)
			for i := 0; i < 6; i++ {
				Y[w][i] = X[3*i+w]
			}
		}

		for w := 0; w < 3; w++ {
			for i := 0; i < 12; i++ {
				iF := float64(i)
				for k := 0; k < 6; k++ {
					kF := float64(k)
					y[w][i] += Y[w][k] * math.Cos(math.Pi/24*(2*iF+7)*(2*kF+1))
				}
				y[w][i] *= math.Sin(math.Pi / 12 * (float64(i) + 0.5))
			}
		}
		for i := 0; i < 36; i++ {
			switch {
			case i <= 5:
				f.curr[i] = 0
			case i <= 11:
				f.curr[i] = y[0][i-6]
			case i <= 17:
				f.curr[i] = y[0][i-6] + y[1][i-12]
			case i <= 23:
				f.curr[i] = y[1][i-12] + y[2][i-18]
			case i <= 29:
				f.curr[i] = y[2][i-18]
			default:
				f.curr[i] = 0
			}
		}
	}

	for i := 0; i < 18; i++ {
		x[i] = f.prev[i+18] + f.curr[i]
	}
	f.prev = f.curr
}
