// TODO(u): Evaluate storing the samples (and residuals) during frame audio
// decoding in a buffer allocated for the stream. This buffer would be allocated
// using BlockSize and NChannels from the StreamInfo block, and it could be
// reused in between calls to Next and ParseNext. This should reduce GC
// pressure.

// Package flac provides access to FLAC (Free Lossless Audio Codec) streams.
//
// A brief introduction of the FLAC stream format [1] follows. Each FLAC stream
// starts with a 32-bit signature ("fLaC"), followed by one or more metadata
// blocks, and then one or more audio frames. The first metadata block
// (StreamInfo) describes the basic properties of the audio stream and it is the
// only mandatory metadata block. Subsequent metadata blocks may appear in an
// arbitrary order.
//
// Please refer to the documentation of the meta [2] and the frame [3] packages
// for a brief introduction of their respective formats.
//
//    [1]: https://www.xiph.org/flac/format.html#stream
//    [2]: https://godoc.org/github.com/mewkiz/flac/meta
//    [3]: https://godoc.org/github.com/mewkiz/flac/frame
package flac

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/mewkiz/flac/frame"
	"github.com/mewkiz/flac/meta"
)

// A Stream contains the metadata blocks and provides access to the audio frames
// of a FLAC stream.
//
// ref: https://www.xiph.org/flac/format.html#stream
type Stream struct {
	// The StreamInfo metadata block describes the basic properties of the FLAC
	// audio stream.
	Info *meta.StreamInfo
	// Zero or more metadata blocks.
	Blocks []*meta.Block
	// Underlying io.Reader.
	r io.Reader
	// Underlying io.Closer of file if opened with Open and ParseFile, and nil
	// otherwise.
	c io.Closer
}

// New creates a new Stream for accessing the audio samples of r. It reads and
// parses the FLAC signature and the StreamInfo metadata block, but skips all
// other metadata blocks.
//
// Call Stream.Next to parse the frame header of the next audio frame, and call
// Stream.ParseNext to parse the entire next frame including audio samples.
func New(r io.Reader) (stream *Stream, err error) {
	// Verify FLAC signature and parse the StreamInfo metadata block.
	br := bufio.NewReader(r)
	stream = &Stream{r: br}
	isLast, err := stream.parseStreamInfo()
	if err != nil {
		return nil, err
	}

	// Skip the remaining metadata blocks.
	for !isLast {
		block, err := meta.New(br)
		if err != nil && err != meta.ErrReservedType {
			return stream, err
		}
		err = block.Skip()
		if err != nil {
			return stream, err
		}
		isLast = block.IsLast
	}

	return stream, nil
}

// signature marks the beginning of a FLAC stream.
var signature = []byte("fLaC")

// parseStreamInfo verifies the signature which marks the beginning of a FLAC
// stream, and parses the StreamInfo metadata block. It returns a boolean value
// which specifies if the StreamInfo block was the last metadata block of the
// FLAC stream.
func (stream *Stream) parseStreamInfo() (isLast bool, err error) {
	// Verify FLAC signature.
	r := stream.r
	var buf [4]byte
	_, err = io.ReadFull(r, buf[:])
	if err != nil {
		return false, err
	}
	if !bytes.Equal(buf[:], signature) {
		return false, fmt.Errorf("flac.parseStreamInfo: invalid FLAC signature; expected %q, got %q", signature, buf)
	}

	// Parse StreamInfo metadata block.
	block, err := meta.Parse(r)
	if err != nil {
		return false, err
	}
	si, ok := block.Body.(*meta.StreamInfo)
	if !ok {
		return false, fmt.Errorf("flac.parseStreamInfo: incorrect type of first metadata block; expected *meta.StreamInfo, got %T", si)
	}
	stream.Info = si
	return block.IsLast, nil
}

// Parse creates a new Stream for accessing the metadata blocks and audio
// samples of r. It reads and parses the FLAC signature and all metadata blocks.
//
// Call Stream.Next to parse the frame header of the next audio frame, and call
// Stream.ParseNext to parse the entire next frame including audio samples.
func Parse(r io.Reader) (stream *Stream, err error) {
	// Verify FLAC signature and parse the StreamInfo metadata block.
	br := bufio.NewReader(r)
	stream = &Stream{r: br}
	isLast, err := stream.parseStreamInfo()
	if err != nil {
		return nil, err
	}

	// Parse the remaining metadata blocks.
	for !isLast {
		block, err := meta.Parse(br)
		if err != nil {
			if err != meta.ErrReservedType {
				return stream, err
			}
			// Skip the body of unknown (reserved) metadata blocks, as stated by
			// the specification.
			//
			// ref: https://www.xiph.org/flac/format.html#format_overview
			err = block.Skip()
			if err != nil {
				return stream, err
			}
		}
		stream.Blocks = append(stream.Blocks, block)
		isLast = block.IsLast
	}

	return stream, nil
}

// Open creates a new Stream for accessing the audio samples of path. It reads
// and parses the FLAC signature and the StreamInfo metadata block, but skips
// all other metadata blocks.
//
// Call Stream.Next to parse the frame header of the next audio frame, and call
// Stream.ParseNext to parse the entire next frame including audio samples.
//
// Note: The Close method of the stream must be called when finished using it.
func Open(path string) (stream *Stream, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	stream, err = New(f)
	if err != nil {
		return nil, err
	}
	stream.c = f
	return stream, err
}

// ParseFile creates a new Stream for accessing the metadata blocks and audio
// samples of path. It reads and parses the FLAC signature and all metadata
// blocks.
//
// Call Stream.Next to parse the frame header of the next audio frame, and call
// Stream.ParseNext to parse the entire next frame including audio samples.
//
// Note: The Close method of the stream must be called when finished using it.
func ParseFile(path string) (stream *Stream, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	stream, err = Parse(f)
	stream.c = f
	return stream, err
}

// Close closes the stream if opened through a call to Open or ParseFile, and
// performs no operation otherwise.
func (stream *Stream) Close() error {
	if r, ok := stream.r.(io.Closer); ok {
		return r.Close()
	}
	return nil
}

// Next parses the frame header of the next audio frame. It returns io.EOF to
// signal a graceful end of FLAC stream.
//
// Call Frame.Parse to parse the audio samples of its subframes.
func (stream *Stream) Next() (f *frame.Frame, err error) {
	return frame.New(stream.r)
}

// ParseNext parses the entire next frame including audio samples. It returns
// io.EOF to signal a graceful end of FLAC stream.
func (stream *Stream) ParseNext() (f *frame.Frame, err error) {
	return frame.Parse(stream.r)
}

// TODO(u): Implement a Seek method.
