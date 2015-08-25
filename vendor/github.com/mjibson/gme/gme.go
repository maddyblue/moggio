/*
Package gme decodes game music files.

This package requires cgo and uses code from
http://blargg.8bitalley.com/libs/audio.html#Game_Music_Emu.
*/
package gme

/*
#include "gme.h"
static short short_index(short *s, int i) {
  return s[i];
}
*/
import "C"

import (
	"fmt"
	"io"
	"time"
	"unsafe"
)

const (
	FadeLength = time.Second * 8
)

var (
	// InfoOnly is the sample rate to New if only track information is needed.
	InfoOnly = C.gme_info_only
)

// New opens the file from b with given sample rate.
func New(b []byte, sampleRate int) (*GME, error) {
	var g GME
	data := unsafe.Pointer(&b[0])
	cerror := C.gme_open_data(data, C.long(len(b)), &g.emu, C.int(sampleRate))
	if err := gmeError(cerror); err != nil {
		return nil, err
	}
	return &g, nil
}

// GME decodes game music.
type GME struct {
	emu *_Ctype_struct_Music_Emu
}

type Track struct {
	PlayLength time.Duration

	// Times; negative if unknown.
	// Length is the total length, if specified by file.
	Length      time.Duration
	IntroLength time.Duration
	LoopLength  time.Duration

	System    string
	Game      string
	Song      string
	Author    string
	Copyright string
	Comment   string
	Dumper    string
}

// Tracks returns the number of tracks in the file.
func (g *GME) Tracks() int {
	return int(C.gme_track_count(g.emu))
}

// Track returns information about the n-th track, 0-based.
func (g *GME) Track(track int) (Track, error) {
	var t *_Ctype_struct_gme_info_t
	cerror := C.gme_track_info(g.emu, &t, C.int(track))
	if err := gmeError(cerror); err != nil {
		return Track{}, err
	}
	return Track{
		PlayLength:  time.Duration(t.play_length) * time.Millisecond,
		Length:      time.Duration(t.length) * time.Millisecond,
		IntroLength: time.Duration(t.intro_length) * time.Millisecond,
		LoopLength:  time.Duration(t.loop_length) * time.Millisecond,
		System:      C.GoString(t.system),
		Game:        C.GoString(t.game),
		Song:        C.GoString(t.song),
		Author:      C.GoString(t.game),
		Copyright:   C.GoString(t.copyright),
		Comment:     C.GoString(t.comment),
		Dumper:      C.GoString(t.dumper),
	}, nil
}

// Start initializes the n-th track for playback, 0-based.
func (g *GME) Start(track int) error {
	err := gmeError(C.gme_start_track(g.emu, C.int(track)))
	if err != nil {
		return err
	}
	t, err := g.Track(track)
	if err != nil {
		return err
	}
	C.gme_set_fade(g.emu, C.int(t.PlayLength/time.Millisecond))
	return nil
}

// Played returns the played time of the current track.
func (g *GME) Played() time.Duration {
	return time.Duration(C.gme_tell(g.emu)) * time.Millisecond
}

// Ended returns whether the current track has ended.
func (g *GME) Ended() bool {
	return C.gme_track_ended(g.emu) == 1
}

// Play decodes the next samples into data. Data is populated with two channels
// interleaved.
func (g *GME) Play(data []int16) (err error) {
	b := make([]C.short, len(data))
	datablock := (*C.short)(unsafe.Pointer(&b[0]))
	cerror := C.gme_play(g.emu, C.int(len(b)), datablock)
	if err := gmeError(cerror); err != nil {
		return err
	}
	for i := range data {
		data[i] = int16(C.short_index(datablock, C.int(i)))
	}
	if g.Ended() {
		return io.EOF
	}
	return nil
}

// Close closes the GME file and frees its used memory.
func (g *GME) Close() {
	if g.emu != nil {
		C.gme_delete(g.emu)
		g.emu = nil
	}
}

// Warning returns the last warning produced and clears it.
func (g *GME) Warning() string {
	if g == nil || g.emu == nil {
		return ""
	}
	return C.GoString(C.gme_warning(g.emu))
}

func gmeError(e _Ctype_gme_err_t) error {
	if e == nil {
		return nil
	}
	return fmt.Errorf("gme: %v", C.GoString(e))
}
