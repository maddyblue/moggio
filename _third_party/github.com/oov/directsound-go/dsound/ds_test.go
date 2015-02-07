package dsound

import (
	"math/rand"
	"syscall"
	"testing"
	"time"
)

var (
	user32           = syscall.MustLoadDLL("user32.dll")
	GetDesktopWindow = user32.MustFindProc("GetDesktopWindow")
)

func TestDirectSoundEnumerate(t *testing.T) {
	err := DirectSoundEnumerate(func(guid *GUID, description string, module string) bool {
		t.Log(guid, description, module)
		return true
	})
	if err != nil {
		t.Error(err)
	}
}

func TestDirectSoundCaptureEnumerate(t *testing.T) {
	err := DirectSoundCaptureEnumerate(func(guid *GUID, description string, module string) bool {
		t.Log(guid, description, module)
		return true
	})
	if err != nil {
		t.Error(err)
	}
}

func initDirectSound() *IDirectSound {
	hasDefaultDevice := false
	DirectSoundEnumerate(func(guid *GUID, description string, module string) bool {
		if guid == nil {
			hasDefaultDevice = true
			return false
		}
		return true
	})
	if !hasDefaultDevice {
		return nil
	}

	ds, err := DirectSoundCreate(nil)
	if err != nil {
		panic(err)
	}

	desktopWindow, _, err := GetDesktopWindow.Call()
	err = ds.SetCooperativeLevel(syscall.Handle(desktopWindow), DSSCL_PRIORITY)
	if err != nil {
		panic(err)
	}

	return ds
}

func TestIDirectSound(t *testing.T) {
	ds := initDirectSound()
	if ds == nil {
		t.Skip("No devices.")
	}

	defer ds.Release()

	caps, err := ds.GetCaps()
	if err != nil {
		t.Error(err)
	}
	t.Log(caps)
}

func TestIDirectSoundBufferStatic(t *testing.T) {
	const SampleRate = 44100
	const Bits = 16
	const Channels = 2
	const BytesPerSec = SampleRate * (Channels * Bits / 8)

	ds := initDirectSound()
	if ds == nil {
		t.Skip("No devices.")
	}
	defer ds.Release()

	// primary buffer

	primaryBuf, err := ds.CreateSoundBuffer(&BufferDesc{
		Flags:       DSBCAPS_PRIMARYBUFFER,
		BufferBytes: 0,
		Format:      nil,
	})
	if err != nil {
		t.Error(err)
	}
	bcaps, err := primaryBuf.GetCaps()
	if err != nil {
		t.Error(err)
	}
	t.Log("PrimaryBufferCaps:", bcaps)

	wfmbuf, err := primaryBuf.GetFormatBytes()
	if err != nil {
		t.Error(err)
	}
	t.Log("WaveFormat Bytes:", wfmbuf)

	wfex, err := primaryBuf.GetFormatWaveFormatEx()
	if err != nil {
		t.Error(err)
	}
	t.Log("WaveFormat WaveFormatEx:", wfex)

	wfext, err := primaryBuf.GetFormatWaveFormatExtensible()
	if err != nil {
		t.Error(err)
	}
	t.Log("WaveFormat WaveFormatExtensible:", wfext)

	err = primaryBuf.SetFormatWaveFormatEx(&WaveFormatEx{
		FormatTag:      WAVE_FORMAT_PCM,
		Channels:       Channels,
		SamplesPerSec:  SampleRate,
		BitsPerSample:  Bits,
		BlockAlign:     Channels * Bits / 8,
		AvgBytesPerSec: BytesPerSec,
		ExtSize:        0,
	})
	if err != nil {
		t.Error(err)
	}

	wfex, err = primaryBuf.GetFormatWaveFormatEx()
	if err != nil {
		t.Error(err)
	}
	t.Log("WaveFormat WaveFormatEx:", wfex)

	primaryBuf.Release()

	// secondary buffer

	secondaryBuf, err := ds.CreateSoundBuffer(&BufferDesc{
		Flags:       DSBCAPS_GLOBALFOCUS | DSBCAPS_GETCURRENTPOSITION2 | DSBCAPS_CTRLVOLUME | DSBCAPS_CTRLPAN | DSBCAPS_CTRLFREQUENCY | DSBCAPS_LOCDEFER,
		BufferBytes: BytesPerSec,
		Format: &WaveFormatEx{
			FormatTag:      WAVE_FORMAT_PCM,
			Channels:       Channels,
			SamplesPerSec:  SampleRate,
			BitsPerSample:  Bits,
			BlockAlign:     Channels * Bits / 8,
			AvgBytesPerSec: BytesPerSec,
			ExtSize:        0,
		},
	})
	if err != nil {
		t.Error(err)
	}
	defer secondaryBuf.Release()

	err = secondaryBuf.SetVolume(0)
	if err != nil {
		t.Error(err)
	}

	vol, err := secondaryBuf.GetVolume()
	if err != nil {
		t.Error(err)
	}
	t.Log("Volume:", vol)

	err = secondaryBuf.SetPan(0)
	if err != nil {
		t.Error(err)
	}

	pan, err := secondaryBuf.GetPan()
	if err != nil {
		t.Error(err)
	}
	t.Log("Pan:", pan)

	freq, err := secondaryBuf.GetFrequency()
	if err != nil {
		t.Error(err)
	}
	t.Log("Frequncy:", freq)

	is1, is2, err := secondaryBuf.LockInt16s(0, BytesPerSec, 0)
	if err != nil {
		t.Error(err)
	}
	t.Log("LockInt16s Buf1Len:", len(is1), "Buf2Len:", len(is2))

	// noise fade-in
	p, ld4 := 0.0, float64(len(is1))
	for i := range is1 {
		is1[i] = int16((rand.Float64()*10000 - 5000) * (p / ld4))
		p += 1
	}
	err = secondaryBuf.UnlockInt16s(is1, is2)
	if err != nil {
		t.Error(err)
	}

	err = secondaryBuf.Play(0, DSBPLAY_LOOPING)
	if err != nil {
		t.Error(err)
	}

	status, err := secondaryBuf.GetStatus()
	if err != nil {
		t.Error(err)
	}
	t.Log("Status:", status)

	time.Sleep(time.Second)
}
