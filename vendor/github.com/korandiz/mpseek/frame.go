// mpseek, a library to support seeking MPEG Audio files
// Copyright (C) 2015 KORÁNDI Zoltán <korandi.z@gmail.com>
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

package mpseek

// A header is a 32-bit word with getters for various bit fields. It also has a
// few utility methods, some of which may panic or return an invalid result when
// called on an invalid header.
type header uint32

// syncword returns the value of the syncword field.
func (h header) syncword() int {
	return int(h>>20) & 0xfff
}

// id returns the value of the ID field.
func (h header) id() int {
	return int(h>>19) & 1
}

// layerIndex returns the value of the layer field.
func (h header) layerIndex() int {
	return int(h>>17) & 3
}

// protectionBit returns the value of the protection_bit field.
func (h header) protectionBit() int {
	return int(h>>16) & 1
}

// bitrateIndex returns the value of the bitrate_index field.
func (h header) bitrateIndex() int {
	return int(h>>12) & 15
}

// samplingFreqIndex returns the value of the sampling_frequency field.
func (h header) samplingFreqIndex() int {
	return int(h>>10) & 3
}

// paddingBit returns the value of the padding_bit field.
func (h header) paddingBit() int {
	return int(h>>9) & 1
}

// mode returns the value of the mode field.
func (h header) mode() int {
	return int(h>>6) & 3
}

// valid returns true if and only if h is a valid, non-free format header.
func (h header) valid() bool {
	return h.syncword() == 0xfff &&
		h.id() == 1 &&
		h.layerIndex() != 0 &&
		h.bitrateIndex() != 15 &&
		h.samplingFreqIndex() != 3 &&
		h.bitrateIndex() != 0
}

// layer returns the layer number.
func (h header) layer() int {
	return 4 - h.layerIndex()
}

// bitrateBps returns the bitrate in bps.
func (h header) bitrateBps() int {
	return 1000 * bitrateKBps[h.layer()-1][h.bitrateIndex()-1]
}

// samplingFreqHz returns the sampling frequency in Hz.
func (h header) samplingFreqHz() int {
	return samplingFreqHz[h.samplingFreqIndex()]
}

// frameSize returns the frame size in bytes.
func (h header) frameSize() int {
	size := 12 * h.bitrateBps()
	if h.layer() != 1 {
		size *= 12
	}
	size /= h.samplingFreqHz()
	size += h.paddingBit()
	if h.layer() == 1 {
		size *= 4
	}
	return size
}

// mainDataSize tells how many bytes of main_data this frame contains.
func (h header) mainDataSize() int {
	n := h.frameSize()

	// header
	n -= 4

	// crc_check
	if h.protectionBit() == 0 {
		n -= 2
	}

	// side information
	if h.mode() == 3 {
		n -= 17 // single channel
	} else {
		n -= 32
	}

	return n
}

// samplingFreqHz maps the value of the sampling_frequency field to Hz.
var samplingFreqHz = [3]int{44100, 48000, 32000}

// bitrateKBps tells the bitrate for every combination of layer and
// bitrate_index.
var bitrateKBps = [3][14]int{
	{32, 64, 96, 128, 160, 192, 224, 256, 288, 320, 352, 384, 416, 448},
	{32, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, 384},
	{32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320},
}

// samplesPerFrame maps layer numbers to the number of samples per frame.
var samplesPerFrame = [3]int{384, 1152, 1152}
