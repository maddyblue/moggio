package pulse

/*
#cgo pkg-config: libpulse-simple

#include <stdlib.h>
#include <pulse/simple.h>
*/
import "C"

// BufferAttr holds information about desired data transfer buffer sizes.
// All values are recommended to be initialized to (uint32) - 1,
// which will choose default values depending on the server.
type BufferAttr struct {
	Maxlength uint32 // Playback and Capture, maximum buffer size in bytes
	Tlength   uint32 // Playback-only, target buffer size in bytes
	Prebuf    uint32 // Playback-only, pre-bufferring in bytes
	Minreq    uint32 // Plyback-only, minimum server-client request size in bytes
	Fragsize  uint32 // Capture-only, fragment size in bytes
}

// NewBufferAttr initializes a BufferAttr with values indicating default.
func NewBufferAttr() *BufferAttr {
	return &BufferAttr{
		Maxlength: ^uint32(0),
		Tlength:   ^uint32(0),
		Prebuf:    ^uint32(0),
		Minreq:    ^uint32(0),
		Fragsize:  ^uint32(0),
	}
}

func (b *BufferAttr) toC() *C.pa_buffer_attr {
	if b == nil {
		return nil
	}
	return &C.pa_buffer_attr{
		maxlength: C.uint32_t(b.Maxlength),
		tlength:   C.uint32_t(b.Tlength),
		prebuf:    C.uint32_t(b.Prebuf),
		minreq:    C.uint32_t(b.Minreq),
		fragsize:  C.uint32_t(b.Fragsize),
	}
}
