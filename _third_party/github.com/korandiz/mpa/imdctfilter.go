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

// imdctFilter performs the IMDCT, windowing, and overlap-add steps of Layer III
// decoding.
type imdctFilter [36]float32

// filter transforms the 18 frequency lines in the input array, and writes the
// result in the output array.
func (f *imdctFilter) filter(input, output []float32, typ int) {
	copy(output, f[18:])

	if typ != 2 {
		imdct36(input, f[:])
		for t := 0; t < 36; t++ {
			f[t] *= mdctWindows[typ][t]
		}
	} else {
		var wInput [6]float32
		var wOutput [12]float32
		for t := 0; t < 36; t++ {
			f[t] = 0
		}
		for w := 0; w < 3; w++ {
			for f := 0; f < 6; f++ {
				wInput[f] = input[3*f+w]
			}
			imdct12(wInput[:], wOutput[:])
			for t := 0; t < 12; t++ {
				f[6+6*w+t] += mdctWindows[2][t] * wOutput[t]
			}
		}
	}

	for i := 0; i < 18; i++ {
		output[i] += f[i]
	}
}
