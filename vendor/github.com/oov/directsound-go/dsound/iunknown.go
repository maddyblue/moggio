package dsound

import (
	"strconv"
	"unsafe"
)

type HResult uintptr

const (
	S_OK           = HResult(0x00000000)
	E_NOTIMPL      = HResult(0x80004001)
	E_NOINTERFACE  = HResult(0x80004002)
	E_POINTER      = HResult(0x80004003)
	E_ABORT        = HResult(0x80004004)
	E_FAIL         = HResult(0x80004005)
	E_UNEXPECTED   = HResult(0x8000FFFF)
	E_ACCESSDENIED = HResult(0x80070005)
	E_HANDLE       = HResult(0x80070006)
	E_OUTOFMEMORY  = HResult(0x8007000E)
	E_INVALIDARG   = HResult(0x80070057)
)

func (hr HResult) Error() string {
	switch hr {
	case S_OK:
		return "Operation successful"
	case E_NOTIMPL:
		return "Not implemented"
	case E_NOINTERFACE:
		return "No such interface supported"
	case E_POINTER:
		return "Pointer that is not valid"
	case E_ABORT:
		return "Operation aborted"
	case E_FAIL:
		return "Unspecified failure"
	case E_UNEXPECTED:
		return "Unexpected failure"
	case E_ACCESSDENIED:
		return "General access denied error"
	case E_HANDLE:
		return "Handle that is not valid"
	case E_OUTOFMEMORY:
		return "Failed to allocate necessary memory"
	case E_INVALIDARG:
		return "One or more arguments are not valid"
	}
	return "Unknown HRESULT value: " + strconv.FormatUint(uint64(hr), 10)
}

type IUnknown struct {
	v *iUnknownVTable
}

type iUnknownVTable struct {
	QueryInterface comProc
	AddRef         comProc
	Release        comProc
}

func (unk *IUnknown) QueryInterface(iid *GUID) (*IUnknown, error) {
	var intf *IUnknown
	r, _, _ := unk.v.QueryInterface.Call(
		uintptr(unsafe.Pointer(unk)),
		uintptr(unsafe.Pointer(iid)),
		uintptr(unsafe.Pointer(&intf)),
	)
	if r != 0 {
		return nil, HResult(r)
	}
	return intf, nil
}

func (unk *IUnknown) AddRef() uint32 {
	r, _, _ := unk.v.AddRef.Call(
		uintptr(unsafe.Pointer(unk)),
	)
	return uint32(r)
}

func (unk *IUnknown) Release() uint32 {
	r, _, _ := unk.v.Release.Call(
		uintptr(unsafe.Pointer(unk)),
	)
	return uint32(r)
}
