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

import (
	"errors"
	"io"
	"math"
)

// A Table translates times and sample indexes to byte offsets.
type Table struct {
	layer int
	sf    int // sampling frequency in Hz
	g     int // granularity, in frames
	nf    int // total number of frames
	offs  []int64
	w     []int8
}

// CreateTable reads the input stream through and creates a seek table for it.
// The granularity (time between seek points) will be an integral number of
// frames, and it will be as close to g seconds as possible. If g == 0, a seek
// point will be created for every frame.
//
// The memory footprint of a seek point is on the order of a few bytes.
//
// When a non-nil *Table is returned, it always contains information about the
// part of the stream which could successfully be read and parsed, even if an
// error occured. EOF is not considered as an error.
//
// As it requires reading the entire stream through, constructing the table  may
// take non-trivial time.
func CreateTable(input io.Reader, g float64) (*Table, error) {
	t := new(Table)
	b := buffer{input: input}

	var err1 error

	var frames [11]struct {
		offs int64
		mdb  int
		mds  int
	}

	for {
		h, err := b.synchronize()
		if err != nil {
			err1 = err
			break
		}

		layer, sf := h.layer(), h.samplingFreqHz()

		if t.layer == 0 {
			t.layer, t.sf = layer, sf
			switch g := g * float64(sf) / float64(samplesPerFrame[layer-1]); {
			case math.IsNaN(g):
				t.g = 1
			case g > math.MaxInt32-0.5:
				t.g = math.MaxInt32
			default:
				t.g = int(math.Max(1, g) + 0.5)
			}
			for i := range frames {
				frames[i].offs = b.offs - 4
			}
		}

		if t.layer != layer {
			err1 = errors.New("layer changed")
			break
		} else if t.sf != sf {
			err1 = errors.New("sampling rate changed")
			break
		}

		copy(frames[1:], frames[:])
		frames[0].offs = b.offs - 4
		if t.layer == 3 {
			mdb, err := b.peekMDB(h)
			if err != nil {
				err1 = err
				break
			}
			frames[0].mdb = mdb
			frames[0].mds = h.mainDataSize()
		}

		if t.nf%t.g == 0 {
			switch t.layer {
			case 1:
				// Filterbank delay: 480 samples. Frame length: 384 samples.
				// ceil(480 / 384) == 2
				t.offs = append(t.offs, frames[2].offs)
			case 2:
				// ceil(480 / 1152) == 1
				t.offs = append(t.offs, frames[1].offs)
			case 3:
				// ceil((480 + 576) / 1152) == 1
				i := 1
				for mds := 0; mds < frames[1].mdb && i < len(frames)-1; i++ {
					mds += frames[i+1].mds
				}
				t.offs = append(t.offs, frames[i].offs)
				t.w = append(t.w, int8(i))
			}
		}
		t.nf++

		b.skip(h.frameSize() - 4)
	}

	if t.layer == 0 {
		return nil, errors.New("no syncword found")
	}
	if err1 == io.EOF {
		err1 = nil
	}
	return t, err1
}

// NSamples returns the length of the stream in samples.
func (t *Table) NSamples() int64 {
	if t.layer == 0 {
		panic("uninitialized seek table")
	}
	return int64(t.nf) * int64(samplesPerFrame[t.layer-1])
}

// NFrames returns the length of the stream in frames.
func (t *Table) NFrames() int {
	if t.layer == 0 {
		panic("uninitialized seek table")
	}
	return t.nf
}

// Length returns the length of the stream in seconds.
func (t *Table) Length() float64 {
	return float64(t.NSamples()) / float64(t.sf)
}

// FindSample looks up the latest seek point with a sample index no greater
// than n. Panics if n < 0.
func (t *Table) FindSample(n int64) Result {
	if t.layer == 0 {
		panic("uninitialized seek table")
	}
	if n < 0 {
		panic("seek to negative position")
	}

	spf := samplesPerFrame[t.layer-1]
	frame := int(n / int64(spf))
	index := frame / t.g
	if index >= len(t.offs) {
		index = len(t.offs) - 1
	}
	frame = index * t.g

	var w int
	switch t.layer {
	case 1:
		w = 2
	case 2:
		w = 1
	case 3:
		w = int(t.w[index])
	}
	if frame < w {
		w = frame
	}

	return Result{
		Sample: int64(frame) * int64(spf),
		Time:   float64(frame) * float64(spf) / float64(t.sf),
		Offset: t.offs[index],
		WarmUp: int(w),
	}
}

// FindTime looks up the latest seek point with a timestamp no greater than s
// seconds. Panics if s < 0.
func (t *Table) FindTime(s float64) Result {
	return t.FindSample(int64(s * float64(t.sf)))
}

// SamplingFrequency returns the sampling frequency in Hz.
func (t *Table) SamplingFrequency() int {
	if t.layer == 0 {
		panic("uninitialized seek table")
	}
	return t.sf
}

// SamplesPerFrame returns the number of samples per frame.
func (t *Table) SamplesPerFrame() int {
	if t.layer == 0 {
		panic("uninitialized seek table")
	}
	return samplesPerFrame[t.layer-1]
}

// A Result is the result of a look-up operation. Sample and Time are the sample
// index and the timestamp of the seek point, respectively. Offset is the byte
// offset to seek to, while WarmUp is the number of frames that need to be
// decoded after the seek to actually get to the seek point.
type Result struct {
	Sample int64
	Time   float64
	Offset int64
	WarmUp int
}
