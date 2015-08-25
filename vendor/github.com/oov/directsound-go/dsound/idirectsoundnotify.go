package dsound

import (
	"syscall"
	"unsafe"
)

type IDirectSoundNotify struct {
	v *iDirectSoundNotifyVTable
}

type iDirectSoundNotifyVTable struct {
	iUnknownVTable
	SetNotificationPositions comProc
}

func (dsn *IDirectSoundNotify) QueryInterface(iid *GUID) (*IUnknown, error) {
	return (*IUnknown)(unsafe.Pointer(dsn)).QueryInterface(iid)
}

func (dsn *IDirectSoundNotify) AddRef() uint32 {
	return (*IUnknown)(unsafe.Pointer(dsn)).AddRef()
}

func (dsn *IDirectSoundNotify) Release() uint32 {
	return (*IUnknown)(unsafe.Pointer(dsn)).Release()
}

const (
	DSBNOTIFICATIONS_MAX = 0x00100000
	DSBPN_OFFSETSTOP     = 0xFFFFFFFF
)

type DSBPOSITIONNOTIFY struct {
	Offset      uint32
	EventNotify syscall.Handle
}

func (dsn *IDirectSoundNotify) SetNotificationPositions(positionNotifies []DSBPOSITIONNOTIFY) error {
	return dsResult(dsn.v.SetNotificationPositions.Call(
		uintptr(unsafe.Pointer(dsn)),
		uintptr(len(positionNotifies)),
		uintptr(unsafe.Pointer(&positionNotifies[0])),
	))
}
