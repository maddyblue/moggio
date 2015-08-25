package dsound

import (
	"syscall"
	"unsafe"
)

type IDirectSound struct {
	v *iDirectSoundVTable
}

type iDirectSoundVTable struct {
	iUnknownVTable
	CreateSoundBuffer    comProc
	GetCaps              comProc
	DuplicateSoundBuffer comProc
	SetCooperativeLevel  comProc
	Compact              comProc
	GetSpeakerConfig     comProc
	SetSpeakerConfig     comProc
	Initialize           comProc
}

func (ds *IDirectSound) QueryInterface(iid *GUID) (*IUnknown, error) {
	return (*IUnknown)(unsafe.Pointer(ds)).QueryInterface(iid)
}

func (ds *IDirectSound) AddRef() uint32 {
	return (*IUnknown)(unsafe.Pointer(ds)).AddRef()
}

func (ds *IDirectSound) Release() uint32 {
	return (*IUnknown)(unsafe.Pointer(ds)).Release()
}

type SCL uint32

const (
	DSSCL_NORMAL       = SCL(0x00000001)
	DSSCL_PRIORITY     = SCL(0x00000002)
	DSSCL_EXCLUSIVE    = SCL(0x00000003)
	DSSCL_WRITEPRIMARY = SCL(0x00000004)
)

func (ds *IDirectSound) SetCooperativeLevel(window syscall.Handle, level SCL) error {
	return dsResult(ds.v.SetCooperativeLevel.Call(
		uintptr(unsafe.Pointer(ds)),
		uintptr(window),
		uintptr(level),
	))
}

type Caps struct {
	size                         uint32
	Flags                        uint32
	MinSecondarySampleRate       uint32
	MaxSecondarySampleRate       uint32
	PrimaryBuffers               uint32
	MaxHwMixingAllBuffers        uint32
	MaxHwMixingStaticBuffers     uint32
	MaxHwMixingStreamingBuffers  uint32
	FreeHwMixingAllBuffers       uint32
	FreeHwMixingStaticBuffers    uint32
	FreeHwMixingStreamingBuffers uint32
	MaxHw3DAllBuffers            uint32
	MaxHw3DStaticBuffers         uint32
	MaxHw3DStreamingBuffers      uint32
	FreeHw3DAllBuffers           uint32
	FreeHw3DStaticBuffers        uint32
	FreeHw3DStreamingBuffers     uint32
	TotalHwMemBytes              uint32
	FreeHwMemBytes               uint32
	MaxContigFreeHwMemBytes      uint32
	UnlockTransferRateHwBuffers  uint32
	PlayCpuOverheadSwBuffers     uint32
	Reserved1                    uint32
	Reserved2                    uint32
}

func (ds *IDirectSound) GetCaps() (*Caps, error) {
	var c Caps
	c.size = uint32(unsafe.Sizeof(c))
	err := dsResult(ds.v.GetCaps.Call(
		uintptr(unsafe.Pointer(ds)),
		uintptr(unsafe.Pointer(&c)),
	))
	if err != nil {
		return nil, err
	}
	return &c, nil
}

type BufferDesc struct {
	size            uint32
	Flags           BufferCapsFlag
	BufferBytes     uint32
	Reserved        uint32
	Format          *WaveFormatEx
	GUID3DAlgorithm GUID
}

type BufferCapsFlag uint32

const (
	DSBCAPS_PRIMARYBUFFER       = BufferCapsFlag(0x00000001)
	DSBCAPS_STATIC              = BufferCapsFlag(0x00000002)
	DSBCAPS_LOCHARDWARE         = BufferCapsFlag(0x00000004)
	DSBCAPS_LOCSOFTWARE         = BufferCapsFlag(0x00000008)
	DSBCAPS_CTRL3D              = BufferCapsFlag(0x00000010)
	DSBCAPS_CTRLFREQUENCY       = BufferCapsFlag(0x00000020)
	DSBCAPS_CTRLPAN             = BufferCapsFlag(0x00000040)
	DSBCAPS_CTRLVOLUME          = BufferCapsFlag(0x00000080)
	DSBCAPS_CTRLPOSITIONNOTIFY  = BufferCapsFlag(0x00000100)
	DSBCAPS_CTRLFX              = BufferCapsFlag(0x00000200)
	DSBCAPS_STICKYFOCUS         = BufferCapsFlag(0x00004000)
	DSBCAPS_GLOBALFOCUS         = BufferCapsFlag(0x00008000)
	DSBCAPS_GETCURRENTPOSITION2 = BufferCapsFlag(0x00010000)
	DSBCAPS_MUTE3DATMAXDISTANCE = BufferCapsFlag(0x00020000)
	DSBCAPS_LOCDEFER            = BufferCapsFlag(0x00040000)
	DSBCAPS_TRUEPLAYPOSITION    = BufferCapsFlag(0x00080000)
)

func (ds *IDirectSound) CreateSoundBuffer(bufferDesc *BufferDesc) (*IDirectSoundBuffer, error) {
	var buf *IDirectSoundBuffer
	bufferDesc.size = uint32(unsafe.Sizeof(*bufferDesc))
	err := dsResult(ds.v.CreateSoundBuffer.Call(
		uintptr(unsafe.Pointer(ds)),
		uintptr(unsafe.Pointer(bufferDesc)),
		uintptr(unsafe.Pointer(&buf)),
		0,
	))
	if err != nil {
		return nil, err
	}
	return buf, nil
}
