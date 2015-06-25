/*
Package spc decodes Super Nintendo SPC files.

This package requires cgo and uses code from
http://blargg.8bitalley.com/libs/audio.html#snes_spc.

*/
package spc

/*
#include "spc.h"
static short short_index(short *s, int i) {
  return s[i];
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// SampleRate returns the SPC sample rate.
func SampleRate() int {
	return int(C.spc_sample_rate)
}

// New opens the SPC file from b.
func New(b []byte) (*SPC, error) {
	s := &SPC{
		spc:    C.spc_new(),
		filter: C.spc_filter_new(),
	}
	data := unsafe.Pointer(&b[0])
	cerror := C.spc_load_spc(s.spc, data, C.long(len(b)))
	if err := spcError(cerror); err != nil {
		return nil, err
	}
	C.spc_clear_echo(s.spc)
	C.spc_filter_clear(s.filter)
	return s, nil
}

// SPC is an SPC decoder.
type SPC struct {
	spc    *_Ctype_struct_SNES_SPC
	filter *_Ctype_struct_SPC_Filter
}

// Play decodes the next samples into data.
func (s *SPC) Play(data []int16) (err error) {
	b := make([]C.short, len(data))
	datablock := (*C.spc_sample_t)(unsafe.Pointer(&b[0]))
	sdatablock := (*C.short)(unsafe.Pointer(&b[0]))
	cerror := C.spc_play(s.spc, C.int(len(b)), sdatablock)
	if err := spcError(cerror); err != nil {
		return err
	}
	C.spc_filter_run(s.filter, datablock, C.int(len(b)))
	for i := range data {
		data[i] = int16(C.short_index(sdatablock, C.int(i)))
	}
	return nil
}

// Close closes the SPC file and frees its used memory.
func (s *SPC) Close() {
	if s.spc != nil {
		C.spc_filter_delete(s.filter)
		C.spc_delete(s.spc)
		s.spc = nil
		s.filter = nil
	}
}

func spcError(e _Ctype_spc_err_t) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("spc: %v", C.GoString(e))
}
