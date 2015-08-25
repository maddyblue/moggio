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

// synthesisFilter implements the synthesis subband filter (Figure 3-A.2).
type synthesisFilter [1024]float32

// filter feeds 32 subband samples to the filterbank and computes the next 32
// PCM samples.
func (f *synthesisFilter) filter(x []float32) {
	// Shifting
	copy(f[64:], f[0:960])

	// Matrixing can be carried out efficiently using the discrete cosine
	// transform. For details, see:
	//
	//   K. Konstantinides, "Fast Subband Filtering in MPEG Audio Coding",
	//   IEEE Signal Processing Letters, Vol. 1, N. 2, pp. 26-28, Feb. 1994
	//
	dct32(x)
	for i := 0; i <= 15; i++ {
		f[i] = x[i+16]
		f[32-i] = -f[i]
	}
	f[16] = 0 // :)
	for i := 1; i <= 15; i++ {
		f[48-i] = -x[i]
		f[48+i] = -x[i]
	}
	f[48] = -x[0]

	// Windowing, etc. This is basically the same as the flow chart in the
	// standard, but with unnecessary copying eliminated.
	for j := 0; j < 32; j++ {
		var y float32
		fp, wp := j, j
		for fp < 1024 {
			y += f[fp] * synthesisWindow[wp]
			y += f[fp+96] * synthesisWindow[wp+32]
			fp += 128
			wp += 64
		}
		x[j] = y
	}
}
