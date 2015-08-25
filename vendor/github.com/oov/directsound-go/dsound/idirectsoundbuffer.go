package dsound

import (
	"unsafe"
)

const maxInt = int(^uint32(0) >> 1)

type IDirectSoundBuffer struct {
	v *iDirectSoundBufferVTable
}

type iDirectSoundBufferVTable struct {
	iUnknownVTable
	GetCaps            comProc
	GetCurrentPosition comProc
	GetFormat          comProc
	GetVolume          comProc
	GetPan             comProc
	GetFrequency       comProc
	GetStatus          comProc
	Initialize         comProc
	Lock               comProc
	Play               comProc
	SetCurrentPosition comProc
	SetFormat          comProc
	SetVolume          comProc
	SetPan             comProc
	SetFrequency       comProc
	Stop               comProc
	Unlock             comProc
	Restore            comProc
}

func (dsb *IDirectSoundBuffer) QueryInterface(iid *GUID) (*IUnknown, error) {
	return (*IUnknown)(unsafe.Pointer(dsb)).QueryInterface(iid)
}

func (dsb *IDirectSoundBuffer) QueryInterfaceIDirectSoundNotify() (*IDirectSoundNotify, error) {
	IID_IDirectSoundNotify := &GUID{0xb0210783, 0x89cd, 0x11d0, [...]byte{0xaf, 0x8, 0x0, 0xa0, 0xc9, 0x25, 0xcd, 0x16}}
	unk, err := dsb.QueryInterface(IID_IDirectSoundNotify)
	if err != nil {
		return nil, err
	}
	return (*IDirectSoundNotify)(unsafe.Pointer(unk)), nil
}

func (dsb *IDirectSoundBuffer) AddRef() uint32 {
	return (*IUnknown)(unsafe.Pointer(dsb)).AddRef()
}

func (dsb *IDirectSoundBuffer) Release() uint32 {
	return (*IUnknown)(unsafe.Pointer(dsb)).Release()
}

type BufferCaps struct {
	size               uint32
	Flags              uint32
	BufferBytes        uint32
	UnlockTransferRate uint32
	PlayCpuOverhead    uint32
}

func (dsb *IDirectSoundBuffer) GetCaps() (*BufferCaps, error) {
	var c BufferCaps
	c.size = uint32(unsafe.Sizeof(c))
	err := dsResult(dsb.v.GetCaps.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(unsafe.Pointer(&c)),
	))
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (dsb *IDirectSoundBuffer) GetCurrentPosition() (currentPlayCursor, currentWriteCursor uint32, err error) {
	err = dsResult(dsb.v.GetCurrentPosition.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(unsafe.Pointer(&currentPlayCursor)),
		uintptr(unsafe.Pointer(&currentWriteCursor)),
	))
	return
}

func (dsb *IDirectSoundBuffer) GetFormatBytes() ([]byte, error) {
	var sz uint32
	err := dsResult(dsb.v.GetFormat.Call(
		uintptr(unsafe.Pointer(dsb)),
		0,
		0,
		uintptr(unsafe.Pointer(&sz)),
	))
	if err != nil {
		return nil, err
	}

	buf := make([]byte, sz)
	err = dsResult(dsb.v.GetFormat.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(sz),
		0,
	))
	if err != nil {
		return nil, err
	}

	return buf, nil
}

func (dsb *IDirectSoundBuffer) GetFormatWaveFormatEx() (*WaveFormatEx, error) {
	var wfex WaveFormatEx
	err := dsResult(dsb.v.GetFormat.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(unsafe.Pointer(&wfex)),
		uintptr(unsafe.Sizeof(wfex)),
		0,
	))
	if err != nil {
		return nil, err
	}

	return &wfex, nil
}

func (dsb *IDirectSoundBuffer) GetFormatWaveFormatExtensible() (*WaveFormatExtensible, error) {
	var wfext WaveFormatExtensible
	err := dsResult(dsb.v.GetFormat.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(unsafe.Pointer(&wfext)),
		uintptr(unsafe.Sizeof(wfext)),
		0,
	))
	if err != nil {
		return nil, err
	}

	return &wfext, nil
}

func (dsb *IDirectSoundBuffer) GetVolume() (int32, error) {
	var vol int32
	err := dsResult(dsb.v.GetVolume.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(unsafe.Pointer(&vol)),
	))
	if err != nil {
		return 0, err
	}
	return vol, nil
}

func (dsb *IDirectSoundBuffer) GetPan() (int32, error) {
	var pan int32
	err := dsResult(dsb.v.GetPan.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(unsafe.Pointer(&pan)),
	))
	if err != nil {
		return 0, err
	}
	return pan, nil
}

func (dsb *IDirectSoundBuffer) GetFrequency() (uint32, error) {
	var freq uint32
	err := dsResult(dsb.v.GetFrequency.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(unsafe.Pointer(&freq)),
	))
	if err != nil {
		return 0, err
	}
	return freq, nil
}

type BufferStatus uint32

const (
	DSBSTATUS_PLAYING     = BufferStatus(0x00000001)
	DSBSTATUS_BUFFERLOST  = BufferStatus(0x00000002)
	DSBSTATUS_LOOPING     = BufferStatus(0x00000004)
	DSBSTATUS_LOCHARDWARE = BufferStatus(0x00000008)
	DSBSTATUS_LOCSOFTWARE = BufferStatus(0x00000010)
	DSBSTATUS_TERMINATED  = BufferStatus(0x00000020)
)

func (dsb *IDirectSoundBuffer) GetStatus() (BufferStatus, error) {
	var s BufferStatus
	err := dsResult(dsb.v.GetStatus.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(unsafe.Pointer(&s)),
	))
	if err != nil {
		return 0, err
	}
	return s, nil
}

type BufferLockFlag uint32

const (
	DSBLOCK_FROMWRITECURSOR = BufferLockFlag(0x00000001)
	DSBLOCK_ENTIREBUFFER    = BufferLockFlag(0x00000002)
)

func (dsb *IDirectSoundBuffer) Lock(offset uint32, bytes uint32, flags BufferLockFlag) (ptr1 uintptr, bytes1 uint32, ptr2 uintptr, bytes2 uint32, err error) {
	err = dsResult(dsb.v.Lock.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(offset),
		uintptr(bytes),
		uintptr(unsafe.Pointer(&ptr1)),
		uintptr(unsafe.Pointer(&bytes1)),
		uintptr(unsafe.Pointer(&ptr2)),
		uintptr(unsafe.Pointer(&bytes2)),
		uintptr(flags),
	))
	return
}

func (dsb *IDirectSoundBuffer) LockBytes(offset uint32, bytes uint32, flags BufferLockFlag) ([]byte, []byte, error) {
	var ptr1, ptr2 *[maxInt]byte
	var bytes1, bytes2 uint32
	err := dsResult(dsb.v.Lock.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(offset),
		uintptr(bytes),
		uintptr(unsafe.Pointer(&ptr1)),
		uintptr(unsafe.Pointer(&bytes1)),
		uintptr(unsafe.Pointer(&ptr2)),
		uintptr(unsafe.Pointer(&bytes2)),
		uintptr(flags),
	))
	if err != nil {
		return nil, nil, err
	}

	var buf1, buf2 []byte
	if ptr1 != nil && bytes1 > 0 {
		buf1 = ptr1[:bytes1]
	}
	if ptr2 != nil && bytes2 > 0 {
		buf2 = ptr2[:bytes2]
	}
	return buf1, buf2, nil
}

func (dsb *IDirectSoundBuffer) LockInt16s(offset uint32, bytes uint32, flags BufferLockFlag) ([]int16, []int16, error) {
	var ptr1, ptr2 *[maxInt>>1]int16
	var bytes1, bytes2 uint32
	err := dsResult(dsb.v.Lock.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(offset),
		uintptr(bytes),
		uintptr(unsafe.Pointer(&ptr1)),
		uintptr(unsafe.Pointer(&bytes1)),
		uintptr(unsafe.Pointer(&ptr2)),
		uintptr(unsafe.Pointer(&bytes2)),
		uintptr(flags),
	))
	if err != nil {
		return nil, nil, err
	}

	var buf1, buf2 []int16
	if ptr1 != nil && bytes1 > 0 {
		buf1 = ptr1[:bytes1>>1]
	}
	if ptr2 != nil && bytes2 > 0 {
		buf2 = ptr2[:bytes2>>1]
	}
	return buf1, buf2, nil
}

type BufferPlayFlag uint32

const (
	DSBPLAY_LOOPING              = BufferPlayFlag(0x000000001)
	DSBPLAY_LOCHARDWARE          = BufferPlayFlag(0x000000002)
	DSBPLAY_LOCSOFTWARE          = BufferPlayFlag(0x000000004)
	DSBPLAY_TERMINATEBY_TIME     = BufferPlayFlag(0x000000008)
	DSBPLAY_TERMINATEBY_DISTANCE = BufferPlayFlag(0x000000010)
	DSBPLAY_TERMINATEBY_PRIORITY = BufferPlayFlag(0x000000020)
)

func (dsb *IDirectSoundBuffer) Play(priority uint32, flags BufferPlayFlag) error {
	return dsResult(dsb.v.Play.Call(
		uintptr(unsafe.Pointer(dsb)),
		0,
		uintptr(priority),
		uintptr(flags),
	))
}

func (dsb *IDirectSoundBuffer) SetCurrentPosition(newPosition uint32) error {
	return dsResult(dsb.v.SetCurrentPosition.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(newPosition),
	))
}

func (dsb *IDirectSoundBuffer) SetFormatBytes(bytes []byte) error {
	return dsResult(dsb.v.SetFormat.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(unsafe.Pointer(&bytes[0])),
	))
}

func (dsb *IDirectSoundBuffer) SetFormatWaveFormatEx(wfex *WaveFormatEx) error {
	return dsResult(dsb.v.SetFormat.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(unsafe.Pointer(wfex)),
	))
}

func (dsb *IDirectSoundBuffer) SetFormatWaveFormatExtensible(wfext *WaveFormatExtensible) error {
	return dsResult(dsb.v.SetFormat.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(unsafe.Pointer(wfext)),
	))
}

func (dsb *IDirectSoundBuffer) SetVolume(volume int32) error {
	return dsResult(dsb.v.SetVolume.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(volume),
	))
}

func (dsb *IDirectSoundBuffer) SetPan(pan int32) error {
	return dsResult(dsb.v.SetPan.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(pan),
	))
}

func (dsb *IDirectSoundBuffer) SetFrequency(freq uint32) error {
	return dsResult(dsb.v.SetFrequency.Call(
		uintptr(unsafe.Pointer(dsb)),
		uintptr(freq),
	))
}

func (dsb *IDirectSoundBuffer) Stop() error {
	return dsResult(dsb.v.Stop.Call(
		uintptr(unsafe.Pointer(dsb)),
	))
}

func (dsb *IDirectSoundBuffer) Unlock(ptr1 uintptr, bytes1 uint32, ptr2 uintptr, bytes2 uint32) error {
	return dsResult(dsb.v.Unlock.Call(
		uintptr(unsafe.Pointer(dsb)),
		ptr1,
		uintptr(bytes1),
		ptr2,
		uintptr(bytes2),
	))
}

func (dsb *IDirectSoundBuffer) UnlockBytes(buf1 []byte, buf2 []byte) error {
	var ptr1, ptr2 uintptr
	var bytes1, bytes2 uint32
	if buf1 != nil {
		ptr1 = uintptr(unsafe.Pointer(&buf1[0]))
		bytes1 = uint32(len(buf1))
	}
	if buf2 != nil {
		ptr2 = uintptr(unsafe.Pointer(&buf2[0]))
		bytes2 = uint32(len(buf2))
	}
	return dsb.Unlock(ptr1, bytes1, ptr2, bytes2)
}

func (dsb *IDirectSoundBuffer) UnlockInt16s(buf1 []int16, buf2 []int16) error {
	var ptr1, ptr2 uintptr
	var bytes1, bytes2 uint32
	if buf1 != nil {
		ptr1 = uintptr(unsafe.Pointer(&buf1[0]))
		bytes1 = uint32(len(buf1) << 1)
	}
	if buf2 != nil {
		ptr2 = uintptr(unsafe.Pointer(&buf2[0]))
		bytes2 = uint32(len(buf2) << 1)
	}
	return dsb.Unlock(ptr1, bytes1, ptr2, bytes2)
}

func (dsb *IDirectSoundBuffer) Restore() error {
	return dsResult(dsb.v.Restore.Call(
		uintptr(unsafe.Pointer(dsb)),
	))
}
