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

import "io"

// A MalformedStream is returned when the decoder encounters a syntax or
// semantic error it's unable to conceal. Such errors usually leave the decoder
// in an out-of-sync state.
type MalformedStream string

// Error returns the error as a string.
func (err MalformedStream) Error() string {
	return string(err)
}

// A Decoder reads and decodes MPEG-1 Audio data from an input stream. It has
// its own internal buffer, so wrapping the input with a bufio.Reader or similar
// object is unnecessary.
type Decoder struct {
	Input io.Reader

	// Header fields
	layer         int
	protectionBit int
	bitrateIndex  int
	samplingFreq  int
	paddingBit    int
	mode          int
	modeExtension int
	copyright     int
	original      int
	emphasis      int

	// All layers
	stream    bitReader
	nChannels int
	sample    [2][1152]float32
	synth     [2]synthesisFilter

	// Layers I & II
	bound         int
	allocation    [2][32]int
	scalefactor12 [2][32][3]int

	// Layer II
	scfsi2 [2][32]int

	// Layer III
	reservoir           reservoirReader
	mainDataBegin       int
	scfsi3              [2][4]int
	part23Length        [2][2]int
	bigValues           [2][2]int
	globalGain          [2][2]int
	scalefacCompress    [2][2]int
	windowSwitchingFlag [2][2]int
	blockType           [2][2]int
	mixedBlockFlag      [2][2]int
	tableSelect         [2][2][3]int
	subblockGain        [2][2][3]int
	region0Count        [2][2]int
	region1Count        [2][2]int
	preflag             [2][2]int
	scalefacScale       [2][2]int
	count1TableSelect   [2][2]int
	scalefacL           [2][2][21]int
	scalefacS           [2][2][12][3]int
	part2Length         [2][2]int
	huffmanData         [2][2][576]float32
	imdct               [2][32]imdctFilter
}

// Getters

// Layer returns the layer (1, 2, or 3) used for the last decoded frame.
// If DecodeFrame hasn't been called yet, or if the last call failed, the return
// value is undefined.
func (d *Decoder) Layer() int {
	return d.layer
}

// NSamples returns the number of samples per channel (384 or 1152) in the last
// decoded frame.
// If DecodeFrame hasn't been called yet, or if the last call failed, the return
// value is undefined.
func (d *Decoder) NSamples() int {
	if d.layer == 1 {
		return 384
	} else {
		return 1152
	}
}

// Bitrate returns the bitrate of the last decoded frame in bps. For free format
// streams, it returns FreeFormat.
// If DecodeFrame hasn't been called yet, or if the last call failed, the return
// value is undefined.
func (d *Decoder) Bitrate() int {
	return bitrateBps[d.layer][d.bitrateIndex]
}

// SamplingFrequency returns the sampling freqency (32k, 44.1k, or 48k) of the
// last decoded frame in Hz.
// If DecodeFrame hasn't been called yet, or if the last call failed, the return
// value is undefined.
func (d *Decoder) SamplingFrequency() int {
	return samplingFreqHz[d.samplingFreq]
}

// Mode returns the mode used for the last decoded frame. The return value is
// one of ModeStereo, ModeJointStereo, ModeDualChannel, and ModeMono.
// If DecodeFrame hasn't been called yet, or if the last call failed, the return
// value is undefined.
func (d *Decoder) Mode() int {
	return d.mode
}

// NChannels returns the number of channels (1 or 2) of the last decoded frame.
// If DecodeFrame hasn't been called yet, or if the last call failed, the return
// value is undefined.
func (d *Decoder) NChannels() int {
	return d.nChannels
}

// Copyrighted returns true if and only if the bitstream is copyright-protected.
// This information comes directly from the frame header, and it doesn't affect
// the decoding process in any way.
// If DecodeFrame hasn't been called yet, or if the last call failed, the return
// value is undefined.
func (d *Decoder) Copyrighted() bool {
	return d.copyright == 1
}

// Original returns true, if the bitstream is an original, and false, if it's a
// copy.
// This information comes directly from the frame header, and it doesn't affect
// the decoding process in any way.
// If DecodeFrame hasn't been called yet, or if the last call failed, the return
// value is undefined.
func (d *Decoder) Original() bool {
	return d.original == 1
}

// Emphasis returns the type of de-emphasis that shall be used for the last
// decoded frame. The return value is one of EmphNone, Emph5015, EmphUnknown,
// and EmphCCITT.
// This information comes directly from the frame header, and it doesn't affect
// the decoding process in any way.
// If DecodeFrame hasn't been called yet, or if the last call failed, the return
// value is undefined.
func (d *Decoder) Emphasis() int {
	return d.emphasis
}

// ReadSamples reads the samples of channel ch of the last decoded frame into
// dst.
// If len(dst) is greater than the number of samples per channel, the remaining
// elements are left intact. If dst is shorter than the number of samples, then
// only the first len(dst) samples are read.
//
// If ch is invalid (ch < 0 or ch >= NChannels), the data for the 0th (left)
// channel is read. This means it's safe to query "both" channels of a mono
// frame.
//
// If DecodeFrame hasn't been called yet, or if the last call failed, the
// samples are still written into dst, but their value is undefined.
//
// The samples are guaranteed to be in the interval [-1, 1], even after an
// error.
func (d *Decoder) ReadSamples(ch int, dst []float32) {
	if ch < 0 || ch >= d.nChannels {
		ch = 0
	}
	copy(dst, d.sample[ch][0:d.NSamples()])
}

// All layers

// DecodeFrame synchronizes the decoder with the bitstream and decodes the next
// frame found. The resulting PCM samples (and other information about the
// frame) can then be queried.
//
// If Input has changed since the last call, it automatically resets the
// decoder.
//
// DecodeFrame returns an error in the following cases:
//   - If it reaches the EOF before finding a syncword, it returns io.EOF.
//   - If it reaches the EOF in the middle of a frame, it returns
//     io.ErrUnexpectedEOF.
//   - If several consecutive attempts at reading the input return no data and
//     no error either, it returns io.ErrNoProgress.
//   - Any other I/O error is passed on verbatim.
//   - If it encounters a condition where it's absolutely impossible or
//     absolutely pointless to continue (such as an invalid value in the 'layer'
//     header field), DecodeFrame returns a MalformedStream, most likely leaving
//     the decoder in an out-of-sync state.
func (d *Decoder) DecodeFrame() error {
	if d.stream.input != d.Input {
		*d = Decoder{
			Input:     d.Input,
			stream:    bitReader{input: d.Input},
			reservoir: reservoirReader{stream: &d.stream},
		}
	}

	var err error

	if err = d.decodeHeader(); err != nil {
		return err
	}

	if d.protectionBit == 0 {
		// Skip over crc_check.
		for i := 0; i < 2; i++ {
			if _, err = d.stream.readByte(); err == io.EOF {
				return io.ErrUnexpectedEOF
			} else if err != nil {
				return err
			}
		}
	}

	switch d.layer {
	case 1:
		err = d.decodeFrame1()
	case 2:
		err = d.decodeFrame2()
	case 3:
		err = d.decodeFrame3()
	}

	if err == io.EOF {
		return io.ErrUnexpectedEOF
	} else if err != nil {
		return err
	}

	// Any ancillary data is ignored.

	d.synthetizeOutput()

	return nil
}

// decodeHeader reads and parses the header of the next frame.
func (d *Decoder) decodeHeader() error {
	header, err := d.findHeader()
	if err != nil {
		return err
	}

	d.layer = 4 - (header>>17)&3
	d.protectionBit = (header >> 16) & 1
	d.bitrateIndex = (header >> 12) & 15
	d.samplingFreq = (header >> 10) & 3
	d.paddingBit = (header >> 9) & 1
	// private_bit is ignored.
	d.mode = (header >> 6) & 3
	d.modeExtension = (header >> 4) & 3
	d.copyright = (header >> 3) & 1
	d.original = (header >> 2) & 1
	d.emphasis = header & 3

	if d.mode == ModeMono {
		d.nChannels = 1
	} else {
		d.nChannels = 2
	}

	return nil
}

// findHeader looks for the next syncword in the input stream and returns the
// header as an integer.
func (d *Decoder) findHeader() (int, error) {
	header := uint32(0)
	consumed := 0

retry:
	for header&0xfff00000 != 0xfff00000 {
		b, err := d.stream.readByte()
		if err != nil {
			if err == io.EOF {
				if header&0xfff000 == 0xfff000 || header&0xfff0 == 0xfff0 {
					err = io.ErrUnexpectedEOF
				}
			}
			return 0, err
		}

		consumed++
		header = header<<8 | uint32(b)
	}

	// Free format headers have a high false positive rate, so we only accept
	// them under some fairly restricted circumstances.
	ffAllowed := consumed == 4 || d.layer != 0 && d.bitrateIndex == 0

	if !d.trueHeader(header, ffAllowed) {
		header &^= 0xff << 24
		goto retry
	}

	return int(header), nil
}

// trueHeader tries to determine if the suspected frame header just found is a
// real one. If ffAllowed is false, all free format headers are considered fake.
func (d *Decoder) trueHeader(header uint32, ffAllowed bool) bool {
	valid, size1 := validateHeader(header)
	if !valid {
		return false
	}

	if size1 == 0 {
		// We don't know where to look for the next frame as this is a free
		// format header. The decision whether to accept it is left to the
		// caller:
		return ffAllowed
	}

	header, ok := d.stream.lookahead(size1 - 4)
	if !ok {
		// The stream is about to end anyway...
		return true
	}

	valid, size2 := validateHeader(header)
	if !valid || size2 == 0 {
		return false
	}

	header, ok = d.stream.lookahead(size1 - 4 + size2)
	if !ok {
		return true
	}

	valid, size3 := validateHeader(header)
	return valid && size3 != 0
}

// validateHeader determines if the parameter is a (syntactically) valid frame
// header, and, if so, it also computes the frame size. For free format frames,
// the reported size is 0.
func validateHeader(header uint32) (valid bool, size int) {
	var (
		sync  = (header >> 20) & 0xfff
		id    = (header >> 19) & 1
		layer = 4 - (header>>17)&3
		br    = (header >> 12) & 15
		sf    = (header >> 10) & 3
		pad   = (header >> 9) & 1
	)

	if sync != 0xfff || id == 0 || layer == 4 || br == 15 || sf == 3 {
		return false, 0
	}

	if br == 0 { // free format
		return true, 0
	}

	size = 12 * bitrateBps[layer][br]
	if layer != 1 {
		size *= 12
	}
	size /= samplingFreqHz[sf]
	size += int(pad)
	if layer == 1 {
		size *= 4
	}

	return true, size
}

// synthetizeOutput feeds the subband samples through the synthesis filterbank
// and clamps the resulting PCM samples to the interval [-1, 1].
func (d *Decoder) synthetizeOutput() {
	samplesPerSubband := 36
	if d.layer == 1 {
		samplesPerSubband = 12
	}

	for ch := 0; ch < d.nChannels; ch++ {
		for s := 0; s < samplesPerSubband; s++ {
			samples := d.sample[ch][32*s : 32*s+32]
			d.synth[ch].filter(samples)
			for i := 0; i < 32; i++ {
				sample := samples[i]
				if sample < -1 {
					sample = -1
				} else if sample > 1 {
					sample = 1
				}
				samples[i] = sample
			}
		}
	}
}

// Layer I

// decodeFrame1 decodes a Layer I frame.
func (d *Decoder) decodeFrame1() error {
	if d.mode == ModeJointStereo {
		d.bound = 4 * (d.modeExtension + 1)
	} else {
		d.bound = 32
	}

	if err := d.decodeAllocation1(); err != nil {
		return err
	}

	if err := d.decodeScalefactors1(); err != nil {
		return err
	}

	if err := d.decodeSamples1(); err != nil {
		return err
	}

	d.dequantize12()

	return nil
}

// decodeAllocation1 decodes the bit allocation data of a Layer I frame.
func (d *Decoder) decodeAllocation1() error {
	for sb := 0; sb < 32; sb++ {
		nch := d.nChannels
		if sb >= d.bound {
			nch = 1
		}

		for ch := 0; ch < nch; ch++ {
			alloc, err := d.stream.readBits(4)
			if err != nil {
				return err
			}

			if alloc == 15 {
				return MalformedStream("alloc == 15")
			}

			if alloc > 0 {
				alloc++
			}

			d.allocation[ch][sb] = alloc
		}
	}

	for sb := d.bound; sb < 32; sb++ {
		d.allocation[1][sb] = d.allocation[0][sb]
	}

	return nil
}

// decodeScalefactors1 decodes the scalefactors of a Layer I frame.
func (d *Decoder) decodeScalefactors1() error {
	for sb := 0; sb < 32; sb++ {
		for ch := 0; ch < d.nChannels; ch++ {
			if d.allocation[ch][sb] != 0 {
				sf, err := d.stream.readBits(6)
				if err != nil {
					return err
				}
				d.scalefactor12[ch][sb][0] = sf
			}
		}
	}
	return nil
}

// decodeSamples1 decodes the samples of a Layer I frame.
func (d *Decoder) decodeSamples1() error {
	for s := 0; s < 12; s++ {
		for sb := 0; sb < 32; sb++ {
			nch := d.nChannels
			if sb >= d.bound {
				nch = 1
			}

			for ch := 0; ch < nch; ch++ {
				if d.allocation[ch][sb] != 0 {
					sample, err := d.stream.readBits(d.allocation[ch][sb])
					if err != nil {
						return err
					}
					d.sample[ch][32*s+sb] = float32(sample)
				}
			}
		}

		for sb := d.bound; sb < 32; sb++ {
			d.sample[1][32*s+sb] = d.sample[0][32*s+sb]
		}
	}
	return nil
}

// dequantize12 dequantizes the samples of a Layer I or Layer II frame.
func (d *Decoder) dequantize12() {
	parts := 1
	if d.layer == 2 {
		parts = 3
	}

	for ch := 0; ch < d.nChannels; ch++ {
		for sb := 0; sb < 32; sb++ {
			alloc := d.allocation[ch][sb]
			if alloc == 0 {
				for s := 0; s < 12*parts; s++ {
					d.sample[ch][32*s+sb] = 0
				}
			} else {
				nb := alloc
				if alloc < 0 {
					nb = bitsPerGroupedSample[-alloc]
				}
				denom := float32(int(1) << uint(nb-1))
				for p := 0; p < parts; p++ {
					sf := scalefactors12[d.scalefactor12[ch][sb][p]]
					for s := 0; s < 12; s++ {
						f := d.sample[ch][32*(12*p+s)+sb]/denom - 1
						if alloc > 0 {
							f = dequantC[nb] * (f + dequantD[nb]) * sf
						} else {
							f = groupedDequantC[nb] * (f + groupedDequantD) * sf
						}
						d.sample[ch][32*(12*p+s)+sb] = f
					}
				}
			}
		}
	}
}

// Layer II

// decodeFrame2 decodes a Layer II frame.
func (d *Decoder) decodeFrame2() error {
	if d.mode == ModeJointStereo {
		d.bound = 4 * (d.modeExtension + 1)
	} else {
		d.bound = 32
	}

	if err := d.decodeAllocation2(); err != nil {
		return err
	}

	if err := d.decodeScalefactorSelectionInformation2(); err != nil {
		return err
	}

	if err := d.decodeScalefactors2(); err != nil {
		return err
	}

	if err := d.decodeSamples2(); err != nil {
		return err
	}

	d.dequantize12()

	return nil
}

// decodeAllocation2 decodes the bit allocation data of a Layer II frame.
func (d *Decoder) decodeAllocation2() error {
	aTab := allocationTables[d.nChannels-1][d.samplingFreq][d.bitrateIndex]
	if aTab == nil {
		return MalformedStream("illegal combination of bitrate and mode")
	}

	for sb := 0; sb < 32; sb++ {
		nch := d.nChannels
		if sb >= d.bound {
			nch = 1
		}

		for ch := 0; ch < nch; ch++ {
			index, err := d.stream.readBits(aTab[sb].nbal)
			if err != nil {
				return err
			}
			d.allocation[ch][sb] = int(aTab[sb].bits[index])
		}

	}
	for sb := d.bound; sb < 32; sb++ {
		d.allocation[1][sb] = d.allocation[0][sb]
	}

	return nil
}

// decodeScalefactorSelectionInformation2 decodes the scalefactor selection
// information of a Layer II frame.
func (d *Decoder) decodeScalefactorSelectionInformation2() error {
	for sb := 0; sb < 32; sb++ {
		for ch := 0; ch < d.nChannels; ch++ {
			if d.allocation[ch][sb] != 0 {
				scfsi, err := d.stream.readBits(2)
				if err != nil {
					return err
				}
				d.scfsi2[ch][sb] = scfsi
			}
		}
	}

	return nil
}

// decodeScalefactors2 decodes the scalefactors of a Layer II frame.
func (d *Decoder) decodeScalefactors2() error {
	for sb := 0; sb < 32; sb++ {
		for ch := 0; ch < d.nChannels; ch++ {
			if d.allocation[ch][sb] != 0 {
				sf, err := d.stream.readBits(6)
				if err != nil {
					return err
				}
				d.scalefactor12[ch][sb][0] = sf

				if d.scfsi2[ch][sb] == 0 || d.scfsi2[ch][sb] == 3 {
					if sf, err = d.stream.readBits(6); err != nil {
						return err
					}
				}
				d.scalefactor12[ch][sb][1] = sf

				if d.scfsi2[ch][sb] <= 1 {
					if sf, err = d.stream.readBits(6); err != nil {
						return err
					}
				}
				d.scalefactor12[ch][sb][2] = sf
			}
		}
	}

	return nil
}

// decodeSamples2 decodes the samples of a Layer II frame.
func (d *Decoder) decodeSamples2() error {
	for gr := 0; gr < 12; gr++ {
		for sb := 0; sb < 32; sb++ {
			nch := d.nChannels
			if sb >= d.bound {
				nch = 1
			}

			for ch := 0; ch < nch; ch++ {
				if alloc := d.allocation[ch][sb]; alloc > 0 {
					for s := 0; s < 3; s++ {
						sample, err := d.stream.readBits(alloc)
						if err != nil {
							return err
						}
						d.sample[ch][32*(3*gr+s)+sb] = float32(sample)
					}
				} else if alloc < 0 {
					sample, err := d.stream.readBits(-alloc)
					if err != nil {
						return err
					}
					nLevels := levelsPerGroupedSample[-alloc]
					d.sample[ch][32*(3*gr+0)+sb] = float32(sample % nLevels)
					sample /= nLevels
					d.sample[ch][32*(3*gr+1)+sb] = float32(sample % nLevels)
					d.sample[ch][32*(3*gr+2)+sb] = float32(sample / nLevels)
				}
			}
		}

		for sb := d.bound; sb < 32; sb++ {
			d.sample[1][32*(3*gr+0)+sb] = d.sample[0][32*(3*gr+0)+sb]
			d.sample[1][32*(3*gr+1)+sb] = d.sample[0][32*(3*gr+1)+sb]
			d.sample[1][32*(3*gr+2)+sb] = d.sample[0][32*(3*gr+2)+sb]
		}
	}

	return nil
}

// Layer III

// decodeFrame3 decodes a Layer III frame.
func (d *Decoder) decodeFrame3() error {
	if err := d.decodeSideInformation3(); err != nil {
		return err
	}

	if err := d.updateReservoir3(); err != nil {
		return err
	}

	// If window_switching_flag == 1 and block_type == 0, we only report the
	// error after reading the side information through and updating the bit
	// reservoir, so the decoder is more likely to stay in sync.
	for gr := 0; gr < 2; gr++ {
		for ch := 0; ch < d.nChannels; ch++ {
			if d.windowSwitchingFlag[gr][ch] == 1 && d.blockType[gr][ch] == 0 {
				return MalformedStream("block_type == 0")
			}
		}
	}

	if err := d.decodeMainData3(); err != nil {
		return err
	}

	// Parsing done, now do the math:
	d.dequantize3()
	d.stereo3()
	d.reorder3()
	d.antialias3()
	d.imdctFilter3()

	return nil
}

// decodeSideInformation3 decodes the side information of a Layer III frame.
func (d *Decoder) decodeSideInformation3() error {
	var err error

	if d.mainDataBegin, err = d.stream.readBits(9); err != nil {
		return err
	}

	// private_bits (ignored)
	if d.mode == ModeMono {
		_, err = d.stream.readBits(5)
	} else {
		_, err = d.stream.readBits(3)
	}
	if err != nil {
		return err
	}

	for ch := 0; ch < d.nChannels; ch++ {
		for scfsiBand := 0; scfsiBand < 4; scfsiBand++ {
			if d.scfsi3[ch][scfsiBand], err = d.stream.readBits(1); err != nil {
				return err
			}
		}
	}

	for gr := 0; gr < 2; gr++ {
		for ch := 0; ch < d.nChannels; ch++ {
			d.part23Length[gr][ch], err = d.stream.readBits(12)
			if err != nil {
				return err
			}

			d.bigValues[gr][ch], err = d.stream.readBits(9)
			if err != nil {
				return err
			}

			d.globalGain[gr][ch], err = d.stream.readBits(8)
			if err != nil {
				return err
			}

			d.scalefacCompress[gr][ch], err = d.stream.readBits(4)
			if err != nil {
				return err
			}

			d.windowSwitchingFlag[gr][ch], err = d.stream.readBits(1)
			if err != nil {
				return err
			}

			if d.windowSwitchingFlag[gr][ch] == 1 {
				d.blockType[gr][ch], err = d.stream.readBits(2)
				if err != nil {
					return err
				}

				d.mixedBlockFlag[gr][ch], err = d.stream.readBits(1)
				if err != nil {
					return err
				}

				// We don't set region0_count and region1_count to their default
				// values here. Instead, we take window_switching_flag into
				// account when decoding the Huffman data.

				for region := 0; region < 2; region++ {
					d.tableSelect[gr][ch][region], err = d.stream.readBits(5)
					if err != nil {
						return err
					}
				}

				for window := 0; window < 3; window++ {
					d.subblockGain[gr][ch][window], err = d.stream.readBits(3)
					if err != nil {
						return err
					}
				}
			} else {
				d.blockType[gr][ch] = 0

				for region := 0; region < 3; region++ {
					d.tableSelect[gr][ch][region], err = d.stream.readBits(5)
					if err != nil {
						return err
					}
				}

				d.region0Count[gr][ch], err = d.stream.readBits(4)
				if err != nil {
					return err
				}

				d.region1Count[gr][ch], err = d.stream.readBits(3)
				if err != nil {
					return err
				}
			}

			d.preflag[gr][ch], err = d.stream.readBits(1)
			if err != nil {
				return err
			}

			d.scalefacScale[gr][ch], err = d.stream.readBits(1)
			if err != nil {
				return err
			}

			d.count1TableSelect[gr][ch], err = d.stream.readBits(1)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// updateReservoir3 moves all main data between the byte pointed by
// main_data_begin and the last byte of the current frame to the bit reservoir.
func (d *Decoder) updateReservoir3() error {
	err1 := d.reservoir.setSize(d.mainDataBegin)

	// Even if setSize failed due to not enough main data being available, we
	// only report the error after loading the main data from the current frame,
	// as it may be required for decoding the next frame.

	if d.bitrateIndex != 0 {
		// Frame size
		mainDataLength := 144 * bitrateBps[d.layer][d.bitrateIndex]
		mainDataLength /= samplingFreqHz[d.samplingFreq]
		mainDataLength += d.paddingBit
		// Header
		mainDataLength -= 4
		// crc_check
		if d.protectionBit == 0 {
			mainDataLength -= 2
		}
		// Side information
		if d.mode == ModeMono {
			mainDataLength -= 17
		} else {
			mainDataLength -= 32
		}

		if err := d.reservoir.load(mainDataLength); err != nil {
			return err
		}
	} else {
		if err := d.reservoir.loadUntilSyncword(); err != nil {
			return err
		}
	}

	return err1
}

// decodeMainData3 decodes the main data (scalefactors & Huffman data) of a
// Layer III frame.
func (d *Decoder) decodeMainData3() error {
	for gr := 0; gr < 2; gr++ {
		for ch := 0; ch < d.nChannels; ch++ {
			if err := d.decodeScalefactors3(gr, ch); err != nil {
				return err
			}
			if err := d.decodeHuffmanData3(gr, ch); err != nil {
				return err
			}
		}
	}
	return nil
}

// decodeScalefactors3 decodes the scalefactors of a Layer III frame. It also
// computes part2_length.
func (d *Decoder) decodeScalefactors3(gr, ch int) error {
	var err error
	s1 := slen1[d.scalefacCompress[gr][ch]]
	s2 := slen2[d.scalefacCompress[gr][ch]]

	if d.blockType[gr][ch] == 2 {
		shortLow := 0
		if d.mixedBlockFlag[gr][ch] == 1 {
			d.part2Length[gr][ch] = 17*s1 + 18*s2
			for sfb := 0; sfb < 8; sfb++ {
				d.scalefacL[gr][ch][sfb], err = d.reservoir.readBits(s1)
				if err != nil {
					return err
				}
			}
			shortLow = 3
		} else {
			d.part2Length[gr][ch] = 18*s1 + 18*s2
		}

		for sfb := shortLow; sfb < 12; sfb++ {
			s := s1
			if sfb >= 6 {
				s = s2
			}
			for w := 0; w < 3; w++ {
				d.scalefacS[gr][ch][sfb][w], err = d.reservoir.readBits(s)
				if err != nil {
					return err
				}
			}
		}
	} else {
		d.part2Length[gr][ch] = 11*s1 + 10*s2
		for scfsiBand := 0; scfsiBand < 4; scfsiBand++ {
			low, high := scfsiBands[scfsiBand], scfsiBands[scfsiBand+1]
			s := s1
			if scfsiBand >= 2 {
				s = s2
			}
			for sfb := low; sfb < high; sfb++ {
				if gr == 0 || d.scfsi3[ch][scfsiBand] == 0 {
					d.scalefacL[gr][ch][sfb], err = d.reservoir.readBits(s)
					if err != nil {
						return err
					}
				} else {
					d.scalefacL[1][ch][sfb] = d.scalefacL[0][ch][sfb]
					d.part2Length[gr][ch] -= s
				}
			}
		}
	}

	return nil
}

// decodeHuffmanData decodes the Huffman data of a Layer III frame. It reads
// exactly part2_3_length - part2_length bits from the reservoir, skipping the
// stuffing bits at the end of the granule, if necessary.
func (d *Decoder) decodeHuffmanData3(gr, ch int) error {
	var (
		regions [3]int
		N       = 0
		bits    = d.part23Length[gr][ch] - d.part2Length[gr][ch]
	)

	// Compute the region boundaries in the big_values partition
	if d.windowSwitchingFlag[gr][ch] == 0 {
		regions = [3]int{
			d.region0Count[gr][ch] + 1,
			d.region0Count[gr][ch] + d.region1Count[gr][ch] + 2,
			2 * d.bigValues[gr][ch],
		}
		if regions[1] > 22 {
			return MalformedStream("region0_count + region1_count > 20")
		}
		regions[0] = scfBandsL[d.samplingFreq][regions[0]]
		regions[1] = scfBandsL[d.samplingFreq][regions[1]]
	} else {
		regions = [3]int{36, 576, 2 * d.bigValues[gr][ch]}
	}
	if regions[2] > 576 {
		return MalformedStream("big_values too large")
	}
	if regions[1] > regions[2] {
		regions[1] = regions[2]
	}
	if regions[0] > regions[1] {
		regions[0] = regions[1]
	}

	// Decode the big_values partition
	for r := 0; r < 3; r++ {
		htree := huffmanTrees[d.tableSelect[gr][ch][r]]
		lbits := linbits[d.tableSelect[gr][ch][r]]

		if htree == nil {
			if d.tableSelect[gr][ch][r] != 0 {
				return MalformedStream("invalid table_select")
			}
			for ; N < regions[r]; N++ {
				d.huffmanData[gr][ch][N] = 0
			}
			continue
		}

		for ; N < regions[r]; N += 2 {
			code, err := d.reservoir.readCode(htree)
			if err != nil {
				return err
			}
			bits -= code >> 16

			for i := 0; i < 2; i++ {
				value := (code >> uint(8-8*i)) & 0xff
				if value == 15 {
					lin, err := d.reservoir.readBits(lbits)
					if err != nil {
						return err
					}
					value += lin
					bits -= lbits
				}
				// We perform the first step of dequantization here, so we don't
				// need two separate buffers for the Huffman coded integers and
				// the dequantized floats.
				valueFl := pow43[value]
				if value != 0 {
					sign, err := d.reservoir.readBits(1)
					if err != nil {
						return err
					}
					if sign == 1 {
						valueFl = -valueFl
					}
				}
				d.huffmanData[gr][ch][N+i] = valueFl
			}
		}
	}

	// Decode the count1 partition
	htree := htreeA
	if d.count1TableSelect[gr][ch] == 1 {
		htree = htreeB
	}

	// To avoid overindexing huffmanData, we quit the loop when N > 572, i.e.
	// when there's no room for another 4 values.
	for ; bits > 0 && N <= 572; N += 4 {
		code, err := d.reservoir.readCode(htree)
		if err != nil {
			return err
		}
		bits -= code >> 8

		for i := 0; i < 4; i++ {
			value := (code >> uint(3-i)) & 1
			valueFl := float32(value)
			if value != 0 {
				sign, err := d.reservoir.readBits(1)
				if err != nil {
					return err
				}
				if sign == 1 {
					valueFl = -1
				}
			}
			d.huffmanData[gr][ch][N+i] = valueFl
		}
	}

	// rzero
	for ; N < 576; N++ {
		d.huffmanData[gr][ch][N] = 0
	}

	// Skip over any stuffing. We don't need these bits, so we don't care much
	// if the read is successful or not.
	if bits < 0 {
		return MalformedStream("Huffman data overread")
	}
	d.reservoir.readBits(bits)

	return nil
}

// dequantize3 performs the dequantization step of Layer III decoding.
func (d *Decoder) dequantize3() {
	for gr := 0; gr < 2; gr++ {
		for ch := 0; ch < d.nChannels; ch++ {
			longLow, longHigh := 0, 22
			shortLow, shortHigh := 0, 13
			if d.blockType[gr][ch] == 2 {
				if d.mixedBlockFlag[gr][ch] == 1 {
					longHigh = 8
					shortLow = 3
				} else {
					longLow = 22
				}
			} else {
				shortLow = 13
			}

			for sfb := longLow; sfb < longHigh; sfb++ {
				exp := 326 - 210 + d.globalGain[gr][ch]
				if sfb < 21 {
					k := d.scalefacL[gr][ch][sfb]
					k += d.preflag[gr][ch] * pretab[sfb]
					k <<= uint(d.scalefacScale[gr][ch] + 1)
					exp -= k
				}
				q := exp2[exp]
				low := scfBandsL[d.samplingFreq][sfb]
				high := scfBandsL[d.samplingFreq][sfb+1]
				for n := low; n < high; n++ {
					d.huffmanData[gr][ch][n] *= q
				}
			}

			for sfb := shortLow; sfb < shortHigh; sfb++ {
				low := scfBandsS[d.samplingFreq][sfb]
				length := scfBandsS[d.samplingFreq][sfb+1] - low
				for window := 0; window < 3; window++ {
					exp := 326 - 210 + d.globalGain[gr][ch]
					exp -= 8 * d.subblockGain[gr][ch][window]
					if sfb < 12 {
						k := d.scalefacS[gr][ch][sfb][window]
						k <<= uint(d.scalefacScale[gr][ch] + 1)
						exp -= k
					}
					q := exp2[exp]
					winLow := 3*low + window*length
					winHigh := winLow + length
					for n := winLow; n < winHigh; n++ {
						d.huffmanData[gr][ch][n] *= q
					}
				}
			}
		}
	}
}

// stereo3 performs the stereo processing step of Layer III decoding.
func (d *Decoder) stereo3() {
	// The standard is fairly vague on this step, and the reference
	// implementation is garbage. So this function may or may not be correct,
	// but at least it's correct enough to pass the compliance test.

	if d.mode != ModeJointStereo || d.modeExtension == 0 {
		return
	}

	ms, intensity := d.modeExtension&2 != 0, d.modeExtension&1 != 0

	for gr := 0; gr < 2; gr++ {
		longLow, longHigh := 0, 22
		shortLow, shortHigh := 0, 13
		if d.blockType[gr][1] == 2 {
			if d.mixedBlockFlag[gr][1] == 1 {
				longHigh = 8
				shortLow = 3
			} else {
				longLow = 22
			}
		} else {
			shortLow = 13
		}

		zeroPartStartS := [3]int{shortLow, shortLow, shortLow}
		nonzeroShort := false
		for window := 0; window < 3; window++ {
		sfbShort:
			for sfb := shortHigh - 1; sfb >= shortLow; sfb-- {
				low := scfBandsS[d.samplingFreq][sfb]
				length := scfBandsS[d.samplingFreq][sfb+1] - low
				winLow := 3*low + window*length
				winHigh := winLow + length
				for f := winLow; f < winHigh; f++ {
					if d.huffmanData[gr][1][f] != 0 {
						zeroPartStartS[window] = sfb + 1
						nonzeroShort = true
						break sfbShort
					}
				}
			}
		}

		zeroPartStartL := longLow
		if nonzeroShort {
			zeroPartStartL = longHigh
		} else {
		sfbLong:
			for sfb := longHigh - 1; sfb >= longLow; sfb-- {
				low := scfBandsL[d.samplingFreq][sfb]
				high := scfBandsL[d.samplingFreq][sfb+1]
				for f := low; f < high; f++ {
					if d.huffmanData[gr][1][f] != 0 {
						zeroPartStartL = sfb + 1
						break sfbLong
					}
				}
			}
		}

		for window := 0; window < 3; window++ {
			for sfb := shortLow; sfb < shortHigh; sfb++ {
				low := scfBandsS[d.samplingFreq][sfb]
				length := scfBandsS[d.samplingFreq][sfb+1] - low
				winLow := 3*low + window*length
				winHigh := winLow + length
				isPos := sfb
				if isPos == 12 {
					isPos = 11
				}
				isPos = d.scalefacS[gr][1][isPos][window]
				if intensity && sfb >= zeroPartStartS[window] && isPos < 7 {
					L := intensityFactors[isPos]
					R := 1 - L
					for f := winLow; f < winHigh; f++ {
						d.huffmanData[gr][1][f] = R * d.huffmanData[gr][0][f]
						d.huffmanData[gr][0][f] *= L
					}
				} else if ms {
					for f := winLow; f < winHigh; f++ {
						M, S := d.huffmanData[gr][0][f], d.huffmanData[gr][1][f]
						d.huffmanData[gr][0][f] = (M + S) / sqrt2
						d.huffmanData[gr][1][f] = (M - S) / sqrt2
					}
				}
			}
		}

		for sfb := longLow; sfb < longHigh; sfb++ {
			low := scfBandsL[d.samplingFreq][sfb]
			high := scfBandsL[d.samplingFreq][sfb+1]
			isPos := sfb
			if isPos == 21 {
				isPos = 20
			}
			isPos = d.scalefacL[gr][1][isPos]
			if intensity && sfb >= zeroPartStartL && isPos < 7 {
				L := intensityFactors[isPos]
				R := 1 - L
				for f := low; f < high; f++ {
					d.huffmanData[gr][1][f] = R * d.huffmanData[gr][0][f]
					d.huffmanData[gr][0][f] *= L
				}
			} else if ms {
				for f := low; f < high; f++ {
					M, S := d.huffmanData[gr][0][f], d.huffmanData[gr][1][f]
					d.huffmanData[gr][0][f] = (M + S) / sqrt2
					d.huffmanData[gr][1][f] = (M - S) / sqrt2
				}
			}
		}
	}
}

// reorder3 performs the reordering step of Layer III decoding.
func (d *Decoder) reorder3() {
	for gr := 0; gr < 2; gr++ {
		for ch := 0; ch < d.nChannels; ch++ {
			if d.blockType[gr][ch] != 2 {
				continue
			}

			shortLow := 0
			if d.mixedBlockFlag[gr][ch] == 1 {
				shortLow = 3
			}

			tmp := d.huffmanData[gr][ch]
			for sfb := shortLow; sfb < 13; sfb++ {
				low := scfBandsS[d.samplingFreq][sfb]
				length := scfBandsS[d.samplingFreq][sfb+1] - low
				low *= 3
				for w := 0; w < 3; w++ {
					for f := 0; f < length; f++ {
						d.huffmanData[gr][ch][low+3*f+w] = tmp[low+length*w+f]
					}
				}
			}
		}
	}
}

// antialias3 performs the alias reduction step of Layer III decoding.
func (d *Decoder) antialias3() {
	for gr := 0; gr < 2; gr++ {
		for ch := 0; ch < d.nChannels; ch++ {
			sbLimit := 32
			if d.blockType[gr][ch] == 2 {
				if d.mixedBlockFlag[gr][ch] == 1 {
					sbLimit = 2
				} else {
					continue
				}
			}

			for sb := 1; sb < sbLimit; sb++ {
				for i := 0; i < 8; i++ {
					tmp0 := d.huffmanData[gr][ch][18*sb-1-i]
					tmp1 := d.huffmanData[gr][ch][18*sb+i]
					d.huffmanData[gr][ch][18*sb-1-i] = tmp0*cs[i] - tmp1*ca[i]
					d.huffmanData[gr][ch][18*sb+i] = tmp1*cs[i] + tmp0*ca[i]
				}
			}
		}
	}
}

// imdctFilter3 performs the IMDCT filtering step of Layer III decoding.
func (d *Decoder) imdctFilter3() {
	for gr := 0; gr < 2; gr++ {
		for ch := 0; ch < d.nChannels; ch++ {
			var tmp [18]float32
			for sb := 0; sb < 32; sb++ {
				n, t := 18*sb, d.blockType[gr][ch]
				if t != 0 && d.mixedBlockFlag[gr][ch] == 1 && sb < 2 {
					t = 0
				}
				d.imdct[ch][sb].filter(d.huffmanData[gr][ch][n:n+18], tmp[:], t)
				for t := 0; t < 18; t++ {
					if t&sb&1 != 0 {
						tmp[t] *= -1
					}
					d.sample[ch][32*(18*gr+t)+sb] = tmp[t]
				}
			}
		}
	}
}
