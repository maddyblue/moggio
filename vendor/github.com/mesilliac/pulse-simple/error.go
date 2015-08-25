package pulse

import "errors"

/*
#cgo pkg-config: libpulse-simple

#include <pulse/error.h>
*/
import "C"

func errorFromCode(e C.int) error {
	cstr := C.pa_strerror(e)
	return errors.New(C.GoString(cstr))
}
