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

// This file is intentionally left un-fmt'd.

package mpa

import "math"

// All layers

// Named constants for the 'emphasis' header field.
const (
	EmphNone    = 0 // No emphasis
	Emph5015    = 1 // 50/15 µs emphasis
	EmphUnknown = 2 // "reserved"
	EmphCCITT   = 3 // CCITT J.17 emphasis
)

// Named constants for the 'mode' header field.
const (
	ModeStereo      = 0
	ModeJointStereo = 1
	ModeDualChannel = 2
	ModeMono        = 3
)

// Indexes of the left and right channels.
const (
	ChLeft  = 0
	ChRight = 1
)

// Indicates a free format bitstream, i.e. one not encoded at one of the
// predefined bitrates.
const FreeFormat = 0

// bitrateBps tells the bitrate in bps for every combination of layer and
// bitrate_index.
var bitrateBps = [4][15]int{
	{}, // There's no Layer 0.

	// Layer I
	{
		FreeFormat,  32000,  64000,  96000, 128000, 160000, 192000, 224000,
		    256000, 288000, 320000, 352000, 384000, 416000, 448000,
	},

	// Layer II
	{
		FreeFormat,  32000,  48000,  56000,  64000,  80000,  96000, 112000,
		    128000, 160000, 192000, 224000, 256000, 320000, 384000,
	},

	// Layer III
	{
		FreeFormat,  32000,  40000,  48000,  56000,  64000,  80000,  96000,
		    112000, 128000, 160000, 192000, 224000, 256000, 320000,
	},
}

// samplingFreqHz, indexed with the value of the sampling_frequency header field
// tells the sampling frequency in Hz.
var samplingFreqHz = [3]int{
	44100,
	48000,
	32000,
}

// Layers I & II

// scalefactors12 contains the scalefactors used in Layers I & II (Table 3-B.1).
// Though a scalefactor index of 63 is forbidden, the last entry is still there
// for error concealment.
var scalefactors12 = [64]float32{
	2.0000000000000, 1.5874010519682, 1.2599210498948, 1.0000000000000,
	0.7937005259841, 0.6299605249474, 0.5000000000000, 0.3968502629920,
	0.3149802624737, 0.2500000000000, 0.1984251314960, 0.1574901312368,
	0.1250000000000, 0.0992125657480, 0.0787450656184, 0.0625000000000,
	0.0496062828740, 0.0393725328092, 0.0312500000000, 0.0248031414370,
	0.0196862664046, 0.0156250000000, 0.0124015707185, 0.0098431332023,
	0.0078125000000, 0.0062007853592, 0.0049215666011, 0.0039062500000,
	0.0031003926796, 0.0024607833005, 0.0019531250000, 0.0015501963398,
	0.0012303916502, 0.0009765625000, 0.0007750981699, 0.0006151958251,
	0.0004882812500, 0.0003875490849, 0.0003075979125, 0.0002441406250,
	0.0001937745424, 0.0001537989562, 0.0001220703125, 0.0000968872712,
	0.0000768994781, 0.0000610351562, 0.0000484436356, 0.0000384497390,
	0.0000305175781, 0.0000242218178, 0.0000192248695, 0.0000152587890,
	0.0000121109089, 0.0000096124347, 0.0000076293945, 0.0000060554544,
	0.0000048062173, 0.0000038146972, 0.0000030277272, 0.0000024031086,
	0.0000019073486, 0.0000015138636, 0.0000012015543,               0,
}

// dequantC contains the values of C in the formula used to dequantize ungrouped
// samples in Layer II (2nd column of Table 3-B.4). It's indexed with the number
// of bits per sample. As the numbers coincide with the ones given by the
// formula in Section 2.4.3.2, this table is also used in Layer I. The value at
// index 2 has been added to make this possible. (No samples in Layer II are
// coded on 2 bits.) Indexes 0 and 1 are unused.
var dequantC = [17]float32{
	            0,             0, 1.33333333333, 1.14285714286,
	1.06666666666, 1.03225806452, 1.01587301587, 1.00787401575,
	1.00392156863, 1.00195694716, 1.00097751711, 1.00048851979,
	1.00024420024, 1.00012208522, 1.00006103888, 1.00003051851,
	1.00001525902,
}

// dequantD contains the values of D in the formula used to dequantize ungrouped
// samples in Layer II (3rd column of Table 3-B.4). See dequantC for details.
var dequantD = [17]float32{
	            0,             0, 0.50000000000, 0.25000000000,
	0.12500000000, 0.06250000000, 0.03125000000, 0.01562500000,
	0.00781250000, 0.00390625000, 0.00195312500, 0.00097656250,
	0.00048828125, 0.00024414063, 0.00012207031, 0.00006103516,
	0.00003051758,
}

// Layer II

// groupedDequantC contains the values of C in the formula used to dequantze
// grouped samples in Layer II (2nd column of Table 3-B.4). It's indexed with
// the number of bits per sample, not the number of bits per triplet.
var groupedDequantC = [...]float32{
	2: 1.33333333333,
	3: 1.60000000000,
	4: 1.77777777777,
}

// groupedDequantD is the value of D in the formula used to dequantize grouped
// samples in Layer II (3rd column of Table 3-B.4). It's just a constant rather
// than an array, because for grouped samples, D is always 0.5.
const groupedDequantD = 0.5

// bitsPerGroupedSample maps the number of bits per triplet to the number of
// bits per sample. These numbers aren't mentioned anywhere in the standard, but
// we still need them to make sense of "the first bit of each of the three
// codes". (Which allegedly have to be inverted.)
var bitsPerGroupedSample = [...]int{
	 5: 2,
	 7: 3,
	10: 4,
}

// levelsPerGroupedSample maps the number of bits per triplet to the number of
// quantization steps (1st and 5th columns of Table 3-B.4).
var levelsPerGroupedSample = [...]int{
	 5: 3,
	 7: 5,
	10: 9,
}

// An aTabRow describes one row of a bit allocation table.
type aTabRow struct {
	nbal int
	bits [16]int8
}

// An allocationTable describes an entire bit allocation table. It gives the
// number of bits per sample (or sample code) rather than the number of
// quantazion steps. Grouping is indicated by negative numbers. Rows above
// sblimit have nbal = 0.
type allocationTable [32]aTabRow

// Table 3-B.2a.
var aTabA = &allocationTable{
	{4, [16]int8{0, -5,  3,  4,   5, 6, 7,  8, 9, 10, 11, 12, 13, 14, 15, 16}},
	{4, [16]int8{0, -5,  3,  4,   5, 6, 7,  8, 9, 10, 11, 12, 13, 14, 15, 16}},
	{4, [16]int8{0, -5,  3,  4,   5, 6, 7,  8, 9, 10, 11, 12, 13, 14, 15, 16}},
	{4, [16]int8{0, -5, -7,  3, -10, 4, 5,  6, 7,  8,  9, 10, 11, 12, 13, 16}},
	{4, [16]int8{0, -5, -7,  3, -10, 4, 5,  6, 7,  8,  9, 10, 11, 12, 13, 16}},
	{4, [16]int8{0, -5, -7,  3, -10, 4, 5,  6, 7,  8,  9, 10, 11, 12, 13, 16}},
	{4, [16]int8{0, -5, -7,  3, -10, 4, 5,  6, 7,  8,  9, 10, 11, 12, 13, 16}},
	{4, [16]int8{0, -5, -7,  3, -10, 4, 5,  6, 7,  8,  9, 10, 11, 12, 13, 16}},
	{4, [16]int8{0, -5, -7,  3, -10, 4, 5,  6, 7,  8,  9, 10, 11, 12, 13, 16}},
	{4, [16]int8{0, -5, -7,  3, -10, 4, 5,  6, 7,  8,  9, 10, 11, 12, 13, 16}},
	{4, [16]int8{0, -5, -7,  3, -10, 4, 5,  6, 7,  8,  9, 10, 11, 12, 13, 16}},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{2, [16]int8{0, -5, -7, 16                                              }},
	{2, [16]int8{0, -5, -7, 16                                              }},
	{2, [16]int8{0, -5, -7, 16                                              }},
	{2, [16]int8{0, -5, -7, 16                                              }},
}

// Table 3-B.2b.
var aTabB = &allocationTable{
	{4, [16]int8{0, -5,  3,  4,   5, 6, 7,  8, 9, 10, 11, 12, 13, 14, 15, 16}},
	{4, [16]int8{0, -5,  3,  4,   5, 6, 7,  8, 9, 10, 11, 12, 13, 14, 15, 16}},
	{4, [16]int8{0, -5,  3,  4,   5, 6, 7,  8, 9, 10, 11, 12, 13, 14, 15, 16}},
	{4, [16]int8{0, -5, -7,  3, -10, 4, 5,  6, 7,  8,  9, 10, 11, 12, 13, 16}},
	{4, [16]int8{0, -5, -7,  3, -10, 4, 5,  6, 7,  8,  9, 10, 11, 12, 13, 16}},
	{4, [16]int8{0, -5, -7,  3, -10, 4, 5,  6, 7,  8,  9, 10, 11, 12, 13, 16}},
	{4, [16]int8{0, -5, -7,  3, -10, 4, 5,  6, 7,  8,  9, 10, 11, 12, 13, 16}},
	{4, [16]int8{0, -5, -7,  3, -10, 4, 5,  6, 7,  8,  9, 10, 11, 12, 13, 16}},
	{4, [16]int8{0, -5, -7,  3, -10, 4, 5,  6, 7,  8,  9, 10, 11, 12, 13, 16}},
	{4, [16]int8{0, -5, -7,  3, -10, 4, 5,  6, 7,  8,  9, 10, 11, 12, 13, 16}},
	{4, [16]int8{0, -5, -7,  3, -10, 4, 5,  6, 7,  8,  9, 10, 11, 12, 13, 16}},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{3, [16]int8{0, -5, -7,  3, -10, 4, 5, 16                               }},
	{2, [16]int8{0, -5, -7, 16                                              }},
	{2, [16]int8{0, -5, -7, 16                                              }},
	{2, [16]int8{0, -5, -7, 16                                              }},
	{2, [16]int8{0, -5, -7, 16                                              }},
	{2, [16]int8{0, -5, -7, 16                                              }},
	{2, [16]int8{0, -5, -7, 16                                              }},
	{2, [16]int8{0, -5, -7, 16                                              }},
}

// Table 3-B.2c.
var aTabC = &allocationTable{
	{4, [16]int8{0, -5, -7, -10,  4, 5, 6,  7, 8,  9, 10, 11, 12, 13, 14, 15}},
	{4, [16]int8{0, -5, -7, -10,  4, 5, 6,  7, 8,  9, 10, 11, 12, 13, 14, 15}},
	{3, [16]int8{0, -5, -7, -10,  4, 5, 6,  7                               }},
	{3, [16]int8{0, -5, -7, -10,  4, 5, 6,  7                               }},
	{3, [16]int8{0, -5, -7, -10,  4, 5, 6,  7                               }},
	{3, [16]int8{0, -5, -7, -10,  4, 5, 6,  7                               }},
	{3, [16]int8{0, -5, -7, -10,  4, 5, 6,  7                               }},
	{3, [16]int8{0, -5, -7, -10,  4, 5, 6,  7                               }},
}

// Table 3-B.2d.
var aTabD = &allocationTable{
	{4, [16]int8{0, -5, -7, -10,  4, 5, 6,  7, 8,  9, 10, 11, 12, 13, 14, 15}},
	{4, [16]int8{0, -5, -7, -10,  4, 5, 6,  7, 8,  9, 10, 11, 12, 13, 14, 15}},
	{3, [16]int8{0, -5, -7, -10,  4, 5, 6,  7                               }},
	{3, [16]int8{0, -5, -7, -10,  4, 5, 6,  7                               }},
	{3, [16]int8{0, -5, -7, -10,  4, 5, 6,  7                               }},
	{3, [16]int8{0, -5, -7, -10,  4, 5, 6,  7                               }},
	{3, [16]int8{0, -5, -7, -10,  4, 5, 6,  7                               }},
	{3, [16]int8{0, -5, -7, -10,  4, 5, 6,  7                               }},
	{3, [16]int8{0, -5, -7, -10,  4, 5, 6,  7                               }},
	{3, [16]int8{0, -5, -7, -10,  4, 5, 6,  7                               }},
	{3, [16]int8{0, -5, -7, -10,  4, 5, 6,  7                               }},
	{3, [16]int8{0, -5, -7, -10,  4, 5, 6,  7                               }},
}

// allocationTables tells which bit allocation table shall be used for every
// combination of mode, sampling_frequency, and bitrate_index. The first index
// is the number of channels minus one, the second one is sampling_frequency,
// and the thrird one is bitrate_index. nils indicate invalid combinations.
var allocationTables = [2][3][16]*allocationTable{
	// Single channel mode
	{
		// 44.1 kHz
		{
			aTabB, aTabC, aTabC, aTabA, aTabA, aTabA, aTabB, aTabB,
			aTabB, aTabB, aTabB, nil,   nil,   nil,   nil,   nil,
		},

		// 48 kHz
		{
			aTabA, aTabC, aTabC, aTabA, aTabA, aTabA, aTabA, aTabA,
			aTabA, aTabA, aTabA, nil,   nil,   nil,   nil,   nil,
		},

		// 32 kHz
		{
			aTabB, aTabD, aTabD, aTabA, aTabA, aTabA, aTabB, aTabB,
			aTabB, aTabB, aTabB, nil,   nil,   nil,   nil,   nil,
		},
	},

	// Other modes
	{
		// 44.1 kHz
		{
			aTabB, nil,   nil,   nil,   aTabC, nil,   aTabC, aTabA,
			aTabA, aTabA, aTabB, aTabB, aTabB, aTabB, aTabB, nil,
		},

		// 48 kHz
		{
			aTabA, nil,   nil,   nil,   aTabC, nil,   aTabC, aTabA,
			aTabA, aTabA, aTabA, aTabA, aTabA, aTabA, aTabA, nil,
		},

		// 32 kHz
		{
			aTabB, nil,   nil,   nil,   aTabD, nil,   aTabD, aTabA,
			aTabA, aTabA, aTabB, aTabB, aTabB, aTabB, aTabB, nil,
		},
	},
}

// Layer III

// slen1 is indexed with scalefac_compress, and it tells the value of slen1.
var slen1 = [16]int{0, 0, 0, 0, 3, 1, 1, 1, 2, 2, 2, 3, 3, 3, 4, 4}

// slen2 is indexed with scalefac_compress, and it tells the value of slen2.
var slen2 = [16]int{0, 1, 2, 3, 0, 1, 2, 3, 1, 2, 3, 1, 2, 3, 2, 3}

// scfsiBands lists the boundaries of scalefactor selection information bands.
var scfsiBands = [5]int{0, 6, 11, 16, 21}

// scfBandsL lists the scalefactor band boundaries when long blocks are used.
// (Table 3-B.8.)
var scfBandsL = [3][23]int{
	// 44.1 kHz
	{
		   0,    4,    8,   12,   16,   20,   24,   30,
		  36,   44,   52,   62,   74,   90,  110,  134,
		 162,  196,  238,  288,  342,  418,  576,
	},

	// 48 kHz
	{
		   0,    4,    8,   12,   16,   20,   24,   30,
		  36,   42,   50,   60,   72,   88,  106,  128,
		 156,  190,  230,  276,  330,  384,  576,
	},

	// 32 kHz
	{
		   0,    4,    8,   12,   16,   20,   24,   30,
		  36,   44,   54,   66,   82,  102,  126,  156,
		 194,  240,  296,  364,  448,  550,  576,
	},
}

// scfBandsS lists the scalefactor band boundaries when short blocks are used.
// (Table 3-B.8.)
var scfBandsS = [3][14]int{
	// 44.1 kHz
	{
		   0,    4,    8,   12,   16,   22,   30,   40,
		  52,   66,   84,  106,  136,  192,
	},

	// 48 kHz
	{
		   0,    4,    8,   12,   16,   22,   28,   38,
		  50,   64,   80,  100,  126,  192,
	},

	// 32 kHz
	{
		   0,    4,    8,   12,   16,   22,   30,   42,
		  58,   78,  104,  138,  180,  192,
	},
}

// linbits is indexed with the value of the table_select field, and it tells the
// value of 'linbits' for the corresponding code table.
var linbits = [32]int{
	 0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,  0,
	 1,  2,  3,  4,  6,  8, 10, 13,  4,  5,  6,  7,  8,  9, 11, 13,
}

// pow43 is a look-up table which maps every index i to i^(4/3).
var pow43 [8192+15]float32

func init() {
	for i := 0; i < 8192; i++ {
		pow43[i] = float32(math.Pow(float64(i), 4.0/3.0))
	}

	// Though the values above 8191 are invalid, they can be represented in the
	// bitstream. As a trivial form of error concealment, we replace these
	// values with 8191. This can be achived at zero run-time cost by
	// constructing the look-up table such that it maps all numbers above 8191
	// to 8191^(4/3).

	for i := 8192; i < len(pow43); i++ {
		pow43[i] = pow43[8191]
	}
}

// exp2 is a look-up table which maps every index i to 2^((i-326)/4). This is
// because the exponent in the dequantization formulae is always a quarter
// integer between -326/4 and 45/4.
var exp2 [372]float32

func init() {
	for i := 0; i < 372; i++ {
		exp2[i] = float32(math.Pow(2, (float64(i)-326)/4))
	}
}

// Table 3-B.6.
var pretab = [21]int{
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 3, 3, 3, 2,
}

// intensityFactors is a look-up table for Layer III intensity stereo. It
// contains the factors for the left channel.
var intensityFactors = [7]float32{
	0.000000000000000,
	0.211324865405187,
	0.366025403784439,
	0.500000000000000,
	0.633974596215561,
	0.788675134594813,
	1.000000000000000,
}

// sqrt2 it the square root of 2.
const sqrt2 = 1.4142135623731

// cs contains the coefficients cs_i for alias reduction (Table 3-B.9).
var cs = [8]float32{
	0.857492925712544,
	0.881741997317705,
	0.949628649102733,
	0.983314592491790,
	0.995517816067586,
	0.999160558178148,
	0.999899195244447,
	0.999993155070280,
}

// ca contains the coefficients ca_i for alias reduction (Table 3-B.9).
var ca = [8]float32 {
	-0.51449575542752657,
	-0.47173196856497235,
	-0.31337745420390184,
	-0.18191319961098118,
	-0.09457419252642066,
	-0.04096558288530405,
	-0.01419856857247115,
	-0.00369997467376004,
}

// mdctWindows contains the precomputed values of the window functions used with
// the MDCT filter.
var mdctWindows = [4][36]float32{
	// Type 0 (normal)
	{
		0.0436193873653360, 0.1305261922200516, 0.2164396139381029,
		0.3007057995042731, 0.3826834323650898, 0.4617486132350339,
		0.5372996083468239, 0.6087614290087207, 0.6755902076156601,
		0.7372773368101240, 0.7933533402912352, 0.8433914458128857,
		0.8870108331782216, 0.9238795325112867, 0.9537169507482268,
		0.9762960071199334, 0.9914448613738104, 0.9990482215818578,
		0.9990482215818578, 0.9914448613738104, 0.9762960071199334,
		0.9537169507482269, 0.9238795325112867, 0.8870108331782218,
		0.8433914458128858, 0.7933533402912352, 0.7372773368101241,
		0.6755902076156604, 0.6087614290087209, 0.5372996083468241,
		0.4617486132350339, 0.3826834323650899, 0.3007057995042733,
		0.2164396139381032, 0.1305261922200516, 0.0436193873653361,
	},

	// Type 1 (start)
	{
		0.043619387365336, 0.130526192220052, 0.216439613938103,
		0.300705799504273, 0.382683432365090, 0.461748613235034,
		0.537299608346824, 0.608761429008721, 0.675590207615660,
		0.737277336810124, 0.793353340291235, 0.843391445812886,
		0.887010833178222, 0.923879532511287, 0.953716950748227,
		0.976296007119933, 0.991444861373810, 0.999048221581858,
		1.000000000000000, 1.000000000000000, 1.000000000000000,
		1.000000000000000, 1.000000000000000, 1.000000000000000,
		0.991444861373810, 0.923879532511287, 0.793353340291235,
		0.608761429008721, 0.382683432365090, 0.130526192220052,
		0.000000000000000, 0.000000000000000, 0.000000000000000,
		0.000000000000000, 0.000000000000000, 0.000000000000000,
	},

	// Type 2 (short)
	{
		0.130526192220052, 0.382683432365090, 0.608761429008721,
		0.793353340291235, 0.923879532511287, 0.991444861373810,
		0.991444861373810, 0.923879532511287, 0.793353340291235,
		0.608761429008721, 0.382683432365090, 0.130526192220052,
	},

	// Type 3 (stop)
	{
		0.000000000000000, 0.000000000000000, 0.000000000000000,
		0.000000000000000, 0.000000000000000, 0.000000000000000,
		0.130526192220052, 0.382683432365090, 0.608761429008721,
		0.793353340291235, 0.923879532511287, 0.991444861373810,
		1.000000000000000, 1.000000000000000, 1.000000000000000,
		1.000000000000000, 1.000000000000000, 1.000000000000000,
		0.999048221581858, 0.991444861373810, 0.976296007119933,
		0.953716950748227, 0.923879532511287, 0.887010833178222,
		0.843391445812886, 0.793353340291235, 0.737277336810124,
		0.675590207615660, 0.608761429008721, 0.537299608346824,
		0.461748613235034, 0.382683432365090, 0.300705799504273,
		0.216439613938103, 0.130526192220052, 0.043619387365336,
	},
}
