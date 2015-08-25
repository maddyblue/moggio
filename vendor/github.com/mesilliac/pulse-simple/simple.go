// pulse-simple wraps PulseAudio's Simple API using cgo,
// for easy audio playback and capture via PulseAudio.
//
// Basic usage is to request a playback or capture stream,
// then write bytes to or read bytes from it.
//
// Reading and writing will block until the given byte slice
// is completely consumed or filled, or an error occurs.
//
// The format of the data will be as requested on stream creation.
//
//  ss := pulse.SampleSpec{pulse.SAMPLE_S16LE, 44100, 2}
//  stream, _ := pulse.Playback("my app", "my stream", &ss)
//  defer stream.Free()
//  defer stream.Drain()
//  stream.Write(data)
//
// More example usage can be found in the examples folder.
//
// For more information, see the PulseAudio Simple API documentation at
// http://www.freedesktop.org/software/pulseaudio/doxygen/simple.html
package pulse

/*
#cgo pkg-config: libpulse-simple

#include <stdlib.h>
#include <pulse/simple.h>
*/
import "C"
import "unsafe"

type StreamDirection C.pa_stream_direction_t

const (
	STREAM_NODIRECTION StreamDirection = C.PA_STREAM_NODIRECTION
	STREAM_PLAYBACK    StreamDirection = C.PA_STREAM_PLAYBACK
	STREAM_RECORD      StreamDirection = C.PA_STREAM_RECORD
	STREAM_UPLOAD      StreamDirection = C.PA_STREAM_UPLOAD
)

type Stream struct {
	simple *C.pa_simple
}

// Capture creates a new stream for recording and returns its pointer.
func Capture(clientName, streamName string, spec *SampleSpec) (*Stream, error) {
	return NewStream("", clientName, STREAM_RECORD, "", streamName, spec, nil, nil)
}

// Playback creates a new stream for playback and returns its pointer.
func Playback(clientName, streamName string, spec *SampleSpec) (*Stream, error) {
	return NewStream("", clientName, STREAM_PLAYBACK, "", streamName, spec, nil, nil)
}

func NewStream(
	serverName, clientName string,
	dir StreamDirection,
	deviceName, streamName string,
	spec *SampleSpec,
	cmap *ChannelMap,
	battr *BufferAttr,
) (*Stream, error) {

	s := new(Stream)

	var server *C.char
	if serverName != "" {
		server = C.CString(serverName)
		defer C.free(unsafe.Pointer(server))
	}

	var dev *C.char
	if deviceName != "" {
		dev = C.CString(deviceName)
		defer C.free(unsafe.Pointer(dev))
	}

	name := C.CString(clientName)
	defer C.free(unsafe.Pointer(name))
	stream_name := C.CString(streamName)
	defer C.free(unsafe.Pointer(stream_name))

	var err C.int

	s.simple = C.pa_simple_new(
		server,
		name,
		C.pa_stream_direction_t(dir),
		dev,
		stream_name,
		spec.toC(),
		cmap.toC(),
		battr.toC(),
		&err,
	)

	if err == C.PA_OK {
		return s, nil
	}
	return s, errorFromCode(err)
}

// Stream.Free closes the stream and frees the associated memory.
// The stream becomes invalid after this has been called.
// This should usually be deferred immediately after obtaining a stream.
func (s *Stream) Free() {
	C.pa_simple_free(s.simple)
}

// Stream.Drain blocks until all buffered data has finished playing.
func (s *Stream) Drain() (int, error) {
	var err C.int
	written := C.pa_simple_drain(s.simple, &err)
	if err == C.PA_OK {
		return int(written), nil
	}
	return int(written), errorFromCode(err)
}

// Stream.Flush flushes the playback buffer, discarding any audio therein
func (s *Stream) Flush() (int, error) {
	var err C.int
	flushed := C.pa_simple_flush(s.simple, &err)
	if err == C.PA_OK {
		return int(flushed), nil
	}
	return int(flushed), errorFromCode(err)
}

// Stream.Write writes the given data to the stream,
// blocking until the data has been written.
func (s *Stream) Write(data []byte) (int, error) {
	var err C.int
	written := C.pa_simple_write(
		s.simple,
		unsafe.Pointer(&data[0]),
		C.size_t(len(data)),
		&err,
	)
	if err == C.PA_OK {
		return int(written), nil
	}
	return int(written), errorFromCode(err)
}

// Stream.Read reads data from the stream,
// blocking until it has filled the provided slice.
func (s *Stream) Read(data []byte) (int, error) {
	var err C.int
	written := C.pa_simple_read(
		s.simple,
		unsafe.Pointer(&data[0]),
		C.size_t(len(data)),
		&err,
	)
	if err == C.PA_OK {
		return int(written), nil
	}
	return int(written), errorFromCode(err)
}

// Stream.Latency returns the playback latency in microseconds.
func (s *Stream) Latency() (uint64, error) {
	var err C.int
	lat := C.pa_simple_get_latency(s.simple, &err)
	if err == C.PA_OK {
		return uint64(lat), nil
	}
	return uint64(lat), errorFromCode(err)
}
