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

// imdct4 performs a direct computation of the IMDCT, while imdct12 and imdct36
// use the method described in the following paper:
//
//   H. Shu, X. Bao, Ch. Toumoulin, L. Luo, "Radix-3 Algorithm for the Fast
//   Computation of Forward and Inverse MDCT", IEEE Signal Processing Letters,
//   Vol. 14, N. 2, pp. 93-96, Feb. 2007
//
// The IMDCT with a window size of N is defined as
//
//   x[n] = Sum(k=0...N/2-1) { X[k] * cos(π/(2*N) * (2*n + 1 + N/2) * (2*k+1)) }
//

// imdct36 computes the IMDCT for N = 36.
func imdct36(in, out []float32) {
	var (
		inA, inB, inC    [6]float32
		outA, outB, outC [12]float32
	)

	for k := 0; k < 6; k++ {
		f, g, h := in[k], in[6+k], in[12+k]
		fd, gd, hd := in[5-k], in[11-k], in[17-k]

		inA[k] = f - gd - h
		inB[k] = (2*f+gd+h)*imdct36c[k] + (h-gd)*imdct36s3[k]
		inC[k] = (fd+g)*imdct36c3[k] - (fd-g+2*hd)*imdct36s[k]
	}

	imdct12(inA[:], outA[:])
	imdct12(inB[:], outB[:])
	imdct12(inC[:], outC[:])

	for n := 0; n < 12; n += 2 {
		outC[n] *= -1
	}

	for n := 0; n < 12; n++ {
		out[3*n+1] = outA[n]
		out[3*n] = (outB[n] + outC[n]) / 2
		out[3*n+2] = (outB[n] - outC[n]) / 2
	}
}

// imdct12 computes the IMDCT for N = 12.
func imdct12(in, out []float32) {
	var (
		inA, inB, inC    [2]float32
		outA, outB, outC [4]float32
	)

	for k := 0; k < 2; k++ {
		f, g, h := in[k], in[2+k], in[4+k]
		fd, gd, hd := in[1-k], in[3-k], in[5-k]

		inA[k] = f - gd - h
		inB[k] = (2*f+gd+h)*imdct12c[k] + (h-gd)*imdct12s3[k]
		inC[k] = (fd+g)*imdct12c3[k] - (fd-g+2*hd)*imdct12s[k]
	}

	imdct4(&inA, &outA)
	imdct4(&inB, &outB)
	imdct4(&inC, &outC)

	outC[0] *= -1
	outC[2] *= -1

	for n := 0; n < 4; n++ {
		out[3*n+1] = outA[n]
		out[3*n] = (outB[n] + outC[n]) / 2
		out[3*n+2] = (outB[n] - outC[n]) / 2
	}
}

// imdct4 computes the IMDCT for N = 4.
func imdct4(in *[2]float32, out *[4]float32) {
	tmp0 := in[0]*imdct4c0 + in[1]*imdct4c1
	tmp1 := in[0]*imdct4c1 - in[1]*imdct4c0

	out[0] = tmp0
	out[1] = -tmp0
	out[2] = tmp1
	out[3] = tmp1
}

// imdct36s contains some constants required by imdct36.
// imdct36s[k] = sin(π * (2*k + 1) / 36)
var imdct36s = [6]float32{
	0.0871557427476582,
	0.2588190451025207,
	0.4226182617406994,
	0.5735764363510460,
	0.7071067811865475,
	0.8191520442889917,
}

// imdct36c contains some constants required by imdct36.
// imdct36c[k] = cos(π * (2*k + 1) / 36)
var imdct36c = [6]float32{
	0.996194698091746,
	0.965925826289068,
	0.906307787036650,
	0.819152044288992,
	0.707106781186548,
	0.573576436351046,
}

// imdct36s3 contains some constants required by imdct36.
// imdct36s3[k] = sqrt(3) * sin(π * (2*k + 1) / 36)
var imdct36s3 = [6]float32{
	0.150958174610347,
	0.448287736084027,
	0.731996301541334,
	0.993463529784308,
	1.224744871391589,
	1.418812959832445,
}

// imdct36c3 contains some constants required by imdct36.
// imdct36c3[k] = sqrt(3) * cos(π * (2*k + 1) / 36)
var imdct36c3 = [6]float32{
	1.725459831325642,
	1.673032607475616,
	1.569771134442792,
	1.418812959832445,
	1.224744871391589,
	0.993463529784308,
}

// imdct12s contains some constants required by imdct12.
// imdct12s[k] = sin(π * (2*k + 1) / 12)
var imdct12s = [2]float32{
	0.258819045102521,
	0.707106781186547,
}

// imdct12c contains some constants required by imdct12.
// imdct12c[k] = cos(π * (2*k + 1) / 12)
var imdct12c = [2]float32{
	0.965925826289068,
	0.707106781186548,
}

// imdct12s3 contains some constants required by imdct12.
// imdct12s3[k] = sqrt(3) * sin(π * (2*k + 1) / 12)
var imdct12s3 = [2]float32{
	0.448287736084027,
	1.224744871391589,
}

// imdct12c3 contains some constants required by imdct12.
// imdct12c3[k] = sqrt(3) * cos(π * (2*k + 1) / 12)
var imdct12c3 = [2]float32{
	1.67303260747562,
	1.22474487139159,
}

// Constants required by imdct4.
const (
	imdct4c0 = 0.382683432365090
	imdct4c1 = -0.923879532511287
)
