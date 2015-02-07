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

// All functions in this file use Lee's algorithm to compute the discrete
// cosine transform. For details, see:
//
//   B. G. Lee, "A New Algorithm to Compute The Discrete Cosine Transform",
//   IEEE Transactions on Acoustics, Speech and Signal Processing, Vol. 32,
//   N. 6, pp. 1243-1245, Dec. 1984
//
// The Nth-order DCT is defined by the following formula:
//
//   X[n] = Sum(k=0...N-1) { x[k] * cos(π * (2*k + 1) * n / (2*N)) }
//

// dct32 computes the 32nd-order discrete cosine transform.
func dct32(data []float32) {
	var even, odd [16]float32
	for i := 0; i < 16; i++ {
		even[i] = data[i] + data[31-i]
		odd[i] = (data[i] - data[31-i]) * dct32c[i]
	}

	dct16(&even)
	dct16(&odd)

	for i := 0; i < 15; i++ {
		odd[i] += odd[i+1]
	}

	for i := 0; i < 16; i++ {
		data[2*i] = even[i]
		data[2*i+1] = odd[i]
	}
}

// dct16 computes the 16th-order discrete cosine transform.
func dct16(data *[16]float32) {
	var even, odd [8]float32
	for i := 0; i < 8; i++ {
		even[i] = data[i] + data[15-i]
		odd[i] = (data[i] - data[15-i]) * dct16c[i]
	}

	dct8(&even)
	dct8(&odd)

	for i := 0; i < 7; i++ {
		odd[i] += odd[i+1]
	}

	for i := 0; i < 8; i++ {
		data[2*i] = even[i]
		data[2*i+1] = odd[i]
	}
}

// dct8 computes the 8th-order discrete cosine transform.
func dct8(data *[8]float32) {
	var even, odd [4]float32
	for i := 0; i < 4; i++ {
		even[i] = data[i] + data[7-i]
		odd[i] = (data[i] - data[7-i]) * dct8c[i]
	}

	dct4(&even)
	dct4(&odd)

	for i := 0; i < 3; i++ {
		odd[i] += odd[i+1]
	}

	for i := 0; i < 4; i++ {
		data[2*i] = even[i]
		data[2*i+1] = odd[i]
	}
}

// dct4 computes the 4th-order discrete cosine transform.
func dct4(data *[4]float32) {
	var even, odd [2]float32
	for i := 0; i < 2; i++ {
		even[i] = data[i] + data[3-i]
		odd[i] = (data[i] - data[3-i]) * dct4c[i]
	}

	dct2(&even)
	dct2(&odd)

	odd[0] += odd[1]

	for i := 0; i < 2; i++ {
		data[2*i] = even[i]
		data[2*i+1] = odd[i]
	}
}

// dct2 computes the 2nd-order discrete cosine transform.
func dct2(data *[2]float32) {
	even := data[0] + data[1]
	odd := (data[0] - data[1]) * dct2c

	data[0] = even
	data[1] = odd
}

// dct32c contains the constants required by dct32.
// dct32c[i] = 0.5 / cos(π * (2*i + 1) / 64)
var dct32c = [16]float32{
	0.500602998235196,
	0.505470959897544,
	0.515447309922625,
	0.531042591089784,
	0.553103896034445,
	0.582934968206134,
	0.622504123035665,
	0.674808341455006,
	0.744536271002299,
	0.839349645415527,
	0.972568237861961,
	1.169439933432885,
	1.484164616314166,
	2.057781009953413,
	3.407608418468719,
	10.190008123548033,
}

// dct16c contains the constants required by dct16.
// dct16c[i] = 0.5 / cos(π * (2*i + 1) / 32)
var dct16c = [8]float32{
	0.502419286188156,
	0.522498614939689,
	0.566944034816358,
	0.646821783359990,
	0.788154623451250,
	1.060677685990347,
	1.722447098238334,
	5.101148618689155,
}

// dct8c contains the constants required by dct8.
// dct8c[i] = 0.5 / cos(π * (2*i + 1) / 16)
var dct8c = [4]float32{
	0.509795579104159,
	0.601344886935045,
	0.899976223136416,
	2.562915447741505,
}

// dct4c contains the constants required by dct4.
// dct4c[i] = 0.5 / cos(π * (2*i + 1) / 8)
var dct4c = [2]float32{
	0.541196100146197,
	1.306562964876376,
}

// dct2c is the single constant required by dct2.
// dct2 = 0.5 / cos(π / 4)
const dct2c = 0.707106781186547
