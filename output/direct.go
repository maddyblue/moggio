// +build windows

package output

import (
	"syscall"
	"unsafe"

	"github.com/mjibson/mog/_third_party/github.com/oov/directsound-go/dsound"
)

var (
	kernel32               = syscall.MustLoadDLL("kernel32")
	CreateEvent            = kernel32.MustFindProc("CreateEventW")
	WaitForMultipleObjects = kernel32.MustFindProc("WaitForMultipleObjects")

	user32           = syscall.MustLoadDLL("user32")
	GetDesktopWindow = user32.MustFindProc("GetDesktopWindow")
)

const (
	numBlock = 8
	bits     = 16
)

const (
	WAIT_OBJECT_0  = 0x00000000
	WAIT_ABANDONED = 0x00000080
	WAIT_TIMEOUT   = 0x00000102
)

type output struct {
	ch      chan float32
	stopped bool

	ds          *dsound.IDirectSound
	sr, chans   int
	buf1, buf2  *dsound.IDirectSoundBuffer
	blockSize   uint32
	blockAlign  int
	bytesPerSec int

	offset uint32
}

func get(sampleRate, channels int) (Output, error) {
	var err error
	o := output{
		sr:    sampleRate,
		chans: channels,
		ch:    make(chan float32, 4096*4),
	}

	o.ds, err = dsound.DirectSoundCreate(nil)
	if err != nil {
		panic(err)
	}
	desktopWindow, _, err := GetDesktopWindow.Call()
	err = o.ds.SetCooperativeLevel(syscall.Handle(desktopWindow), dsound.DSSCL_PRIORITY)
	if err != nil {
		panic(err)
	}
	o.buf1, err = o.ds.CreateSoundBuffer(&dsound.BufferDesc{
		Flags: dsound.DSBCAPS_PRIMARYBUFFER,
	})
	if err != nil {
		panic(err)
	}
	o.blockAlign = channels * bits / 8
	o.bytesPerSec = sampleRate * o.blockAlign
	o.blockSize = uint32((sampleRate / numBlock) * o.blockAlign)
	format := &dsound.WaveFormatEx{
		FormatTag:      dsound.WAVE_FORMAT_PCM,
		Channels:       uint16(channels),
		SamplesPerSec:  uint32(sampleRate),
		BitsPerSample:  bits,
		BlockAlign:     uint16(o.blockAlign),
		AvgBytesPerSec: uint32(o.bytesPerSec),
	}
	if err = o.buf1.SetFormatWaveFormatEx(format); err != nil {
		panic(err)
	}
	o.buf1.Release()
	o.buf2, err = o.ds.CreateSoundBuffer(&dsound.BufferDesc{
		Flags:       dsound.DSBCAPS_GLOBALFOCUS | dsound.DSBCAPS_CTRLPOSITIONNOTIFY,
		BufferBytes: o.blockSize * numBlock,
		Format:      format,
	})
	if err != nil {
		panic(err)
	}

	go o.start()
	return &o, nil
}

func (o *output) Push(samples []float32) {
	for _, s := range samples {
		o.ch <- s
	}
}

func (o *output) start() {
	notifies := make([]dsound.DSBPOSITIONNOTIFY, numBlock)
	events := make([]syscall.Handle, 0)
	for i := range notifies {
		h, _, _ := CreateEvent.Call(0, 0, 0, 0)
		notifies[i].EventNotify = syscall.Handle(h)
		notifies[i].Offset = uint32(i) * o.blockSize
		events = append(events, syscall.Handle(h))
	}

	notif, err := o.buf2.QueryInterfaceIDirectSoundNotify()
	if err != nil {
		panic(err)
	}
	defer notif.Release()

	err = notif.SetNotificationPositions(notifies)
	if err != nil {
		panic(err)
	}

	o.play()

	for {
		r, _, _ := WaitForMultipleObjects.Call(
			uintptr(uint32(len(events))),
			uintptr(unsafe.Pointer(&events[0])),
			0,
			0xFFFFFFFF,
		)
		switch {
		case WAIT_OBJECT_0 <= r && r < WAIT_OBJECT_0+uintptr(len(events)):
			idx := int(r - WAIT_OBJECT_0)
			blockPos := (idx - 1 + numBlock) % numBlock
			o.fill(blockPos)

		case WAIT_ABANDONED <= r && r < WAIT_ABANDONED+uintptr(len(events)):
			panic("wait abandoned")

		case r == WAIT_TIMEOUT:
			panic("wait timeout")
		}
	}
}

func (o *output) fill(block int) {
	b1, b2, err := o.buf2.LockInt16s(uint32(block)*o.blockSize, o.blockSize, 0)
	if err != nil {
		panic(err)
	}
	i16 := make([]int16, len(b1)+len(b2))
Loop:
	for i := range i16 {
		select {
		case s := <-o.ch:
			i16[i] = int16(s * 32767)
		default:
			break Loop
		}
	}
	n := copy(b1, i16)
	i16 = i16[n:]
	n = copy(b2, i16)
	i16 = i16[n:]
	o.buf2.UnlockInt16s(b1, b2)
}

func (o *output) Stop() {
	if err := o.buf2.Stop(); err != nil {
		panic(err)
	}
	o.stopped = true
}

func (o *output) Start() {
	if err := o.buf2.Restore(); err != nil {
		panic(err)
	}
	if o.stopped {
		o.play()
	}
}

func (o *output) play() {
	if err := o.buf2.Play(0, dsound.DSBPLAY_LOOPING); err != nil {
		panic(err)
	}
}
