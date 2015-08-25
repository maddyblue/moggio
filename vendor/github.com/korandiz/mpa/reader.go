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

// Named constants for Reader.Format.
const (
	S16_LE  = iota // Signed, 16-bit, little endian
	S16_BE         // Signed, 16-bit, big endian
	U16_LE         // Unsigned, 16-bit, little endian
	U16_BE         // Unsigned, 16-bit, big endian
	S24_LE         // Signed, 24-bit, little endian
	S24_BE         // Signed, 24-bit, big endian
	U24_LE         // Unsigned, 24-bit, little endian
	U24_BE         // Unsigned, 24-bit, big endian
	S24_3LE        // Signed, packed 24-bit, little endian
	S24_3BE        // Signed, packed 24-bit, big endian
	U24_3LE        // Unsigned, packed 24-bit, little endian
	U24_3BE        // Unsigned, packed 24-bit, big endian
	S32_LE         // Signed, 32-bit, little endian
	S32_BE         // Signed, 32-bit, big endian
	U32_LE         // Unsigned, 32-bit, little endian,
	U32_BE         // Unsigned, 32-bit, big endian
	S8             // Signed, 8-bit
	U8             // Unsigned, 8-bit
	S20_3LE        // Signed, packed 20-bit, little endian
	S20_3BE        // Signed, packed 20-bit, big endian
	U20_3LE        // Unsigned, packed 20-bit, little endian
	U20_3BE        // Unsigned, packed 20-bit, big endian
	S18_3LE        // Signed, packed 18-bit, little endian
	S18_3BE        // Signed, packed 18-bit, big endian
	U18_3LE        // Unsigned, packed 18-bit, little endian
	U18_3BE        // Unsigned, packed 18-bit, big endian
)

// A formatDescription specifies the signedness, endiannes and bit depth of a
// PCM stream.
type formatDescription struct {
	signed bool
	bits   int
	size   int
	le     bool
}

// formatDescriptions lists all formats supported by Reader.
var formatDescriptions = [...]formatDescription{
	S16_LE:  {true, 16, 2, true},
	S16_BE:  {true, 16, 2, false},
	U16_LE:  {false, 16, 2, true},
	U16_BE:  {false, 16, 2, false},
	S24_LE:  {true, 24, 4, true},
	S24_BE:  {true, 24, 4, false},
	U24_LE:  {false, 24, 4, true},
	U24_BE:  {false, 24, 4, false},
	S24_3LE: {true, 24, 3, true},
	S24_3BE: {true, 24, 3, false},
	U24_3LE: {false, 24, 3, true},
	U24_3BE: {false, 24, 3, false},
	S32_LE:  {true, 32, 4, true},
	S32_BE:  {true, 32, 4, false},
	U32_LE:  {false, 32, 4, true},
	U32_BE:  {false, 32, 4, false},
	S8:      {true, 8, 1, false},
	U8:      {false, 8, 1, false},
	S20_3LE: {true, 20, 3, true},
	S20_3BE: {true, 20, 3, false},
	U20_3LE: {false, 20, 3, true},
	U20_3BE: {false, 20, 3, false},
	S18_3LE: {true, 18, 3, true},
	S18_3BE: {true, 18, 3, false},
	U18_3LE: {false, 18, 3, true},
	U18_3BE: {false, 18, 3, false},
}

// A Reader wraps a Decoder and converts its output into a byte stream.
type Reader struct {
	Decoder *Decoder
	Format  int
	Mono    bool
	Swap    bool
	buffer  [8 * 1152]byte
	unread  []byte
}

// Read reads up to len(dst) bytes of PCM data into dst and returns the number
// of bytes read.
func (r *Reader) Read(dst []byte) (int, error) {
	n := 0
	for len(dst) > 0 {
		if len(r.unread) == 0 {
			if err := r.refill(); err != nil {
				return n, err
			}
		}
		m := copy(dst, r.unread)
		n, dst, r.unread = n+m, dst[m:], r.unread[m:]
	}
	return n, nil
}

// refill fills the internal buffer with one frame's worth of data.
func (r *Reader) refill() error {
	if err := r.Decoder.DecodeFrame(); err != nil {
		return err
	}
	var samples [2][1152]float32
	r.Decoder.ReadSamples(0, samples[0][:])
	r.Decoder.ReadSamples(1, samples[1][:])
	r.convert(&samples)
	return nil
}

// convert converts the floating-point samples to the output format, writes the
// result into the buffer, and updates 'unread' accordingly.
func (r *Reader) convert(samples *[2][1152]float32) {
	nSamples := r.Decoder.NSamples()
	nChannels := 2
	if r.Mono {
		nChannels = 1
	}

	format := formatDescriptions[r.Format]
	offset := float32(1)
	if format.signed {
		offset = 0
	}
	multiplier := float32(int(1) << uint(format.bits-1))
	max := (1 + offset) * multiplier
	le := format.le
	size := format.size
	jump := size
	if !r.Mono {
		jump *= 2
	}

	for outCh := 0; outCh < nChannels; outCh++ {
		inCh := outCh
		if r.Swap {
			inCh = 1 - outCh
		}
		ptr := outCh * size

		for t := 0; t < nSamples; t++ {
			s := (samples[inCh][t] + offset) * multiplier
			if s >= max {
				s--
			}
			var si int
			if s < 0 {
				si = int(s - 0.5)
			} else {
				si = int(s + 0.5)
			}
			if le {
				for i := 0; i < size; i++ {
					r.buffer[ptr+i] = byte(si & 0xff)
					si >>= 8
				}
			} else {
				for i := 0; i < size; i++ {
					r.buffer[ptr+size-1-i] = byte(si & 0xff)
					si >>= 8
				}
			}
			ptr += jump
		}

		if r.Mono {
			break
		}
	}

	r.unread = r.buffer[0 : jump*nSamples]
}
