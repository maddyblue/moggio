package vorbis

import (
	"io"

	"github.com/jfreymuth/go-vorbis/ogg"
)

type Vorbis struct {
	reader   *ogg.Reader
	header   identification
	comments comments
	setup    setup
	overlap  [][]float32
	position uint64
}

// Open reads the vorbis headers from an io.Reader and performs the setup needed to decode the data.
func Open(in io.Reader) (*Vorbis, error) {
	return OpenOgg(ogg.NewReader(in))
}

// Open reads the vorbis headers from an ogg.Reader and performs the setup needed to decode the data.
func OpenOgg(in *ogg.Reader) (*Vorbis, error) {
	v := new(Vorbis)
	v.reader = in
	r, err := ogg.NewBitReaderErr(v.reader.NextPacket())
	if err != nil {
		return nil, err
	}
	err = v.header.ReadFrom(r)
	if err != nil {
		return nil, err
	}
	b, err := v.reader.NextPacket()
	if err != nil {
		return nil, err
	}
	v.comments.ReadFrom(b)
	r, err = ogg.NewBitReaderErr(v.reader.NextPacket())
	if err != nil {
		return nil, err
	}
	v.setup.init(&v.header)
	err = v.setup.ReadFrom(r)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// DecodePacket decodes a packet from the input.
// The result is a two-dimensional array, the first index corresponds to the channel, the second to the sample position.
// The number of samples decoded can vary.
func (v *Vorbis) DecodePacket() ([][]float32, error) {
	var out [][]float32
	if v.overlap == nil {
		r, err := ogg.NewBitReaderErr(v.reader.NextPacket())
		if err != nil {
			return nil, err
		}
		out, v.overlap, err = v.setup.decodePacket(r, nil)
		v.position += uint64(len(out[0]))
	}
	r, err := ogg.NewBitReaderErr(v.reader.NextPacket())
	if err != nil {
		return nil, err
	}
	out, v.overlap, err = v.setup.decodePacket(r, v.overlap)
	v.position += uint64(len(out[0]))
	return out, err
}

// SamplePosition returns the number of the first sample that will be returned by the next call to DecodePacket.
func (v *Vorbis) SamplePosition() uint64 {
	return v.position
}

// SampleRate returns the sample rate of the vorbis file
func (v *Vorbis) SampleRate() int {
	return int(v.header.audioSampleRate)
}

// SampleRate returns the number of channels of the vorbis file
func (v *Vorbis) Channels() int {
	return int(v.header.audioChannels)
}

// Bitrate returns an estimated bitrate.
// Some (or all) of the returned values may be 0, meaning that the encoder did not produce an estimate.
func (v *Vorbis) Bitrate() (min, nominal, max int) {
	return int(v.header.bitrateMinimum), int(v.header.bitrateNominal), int(v.header.bitrateMaximum)
}

func (v *Vorbis) VendorString() string {
	return v.comments.vendor
}

func (v *Vorbis) Comments() []string {
	return v.comments.comments
}
