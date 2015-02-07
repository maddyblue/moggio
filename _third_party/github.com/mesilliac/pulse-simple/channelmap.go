package pulse

import (
	"errors"
	"fmt"
	"strings"
	"unsafe"
)

/*
#cgo pkg-config: libpulse-simple

#include <stdlib.h>
#include <pulse/channelmap.h>
*/
import "C"

type ChannelPosition C.pa_channel_position_t

const (
	CHANNEL_POSITION_INVALID ChannelPosition = C.PA_CHANNEL_POSITION_INVALID
	CHANNEL_POSITION_MONO    ChannelPosition = C.PA_CHANNEL_POSITION_MONO

	CHANNEL_POSITION_FRONT_LEFT   ChannelPosition = C.PA_CHANNEL_POSITION_FRONT_LEFT
	CHANNEL_POSITION_FRONT_RIGHT  ChannelPosition = C.PA_CHANNEL_POSITION_FRONT_RIGHT
	CHANNEL_POSITION_FRONT_CENTER ChannelPosition = C.PA_CHANNEL_POSITION_FRONT_CENTER

	CHANNEL_POSITION_LEFT   ChannelPosition = C.PA_CHANNEL_POSITION_LEFT
	CHANNEL_POSITION_RIGHT  ChannelPosition = C.PA_CHANNEL_POSITION_RIGHT
	CHANNEL_POSITION_CENTER ChannelPosition = C.PA_CHANNEL_POSITION_CENTER

	CHANNEL_POSITION_REAR_CENTER ChannelPosition = C.PA_CHANNEL_POSITION_REAR_CENTER
	CHANNEL_POSITION_REAR_LEFT   ChannelPosition = C.PA_CHANNEL_POSITION_REAR_LEFT
	CHANNEL_POSITION_REAR_RIGHT  ChannelPosition = C.PA_CHANNEL_POSITION_REAR_RIGHT

	CHANNEL_POSITION_LFE       ChannelPosition = C.PA_CHANNEL_POSITION_LFE
	CHANNEL_POSITION_SUBWOOFER ChannelPosition = C.PA_CHANNEL_POSITION_SUBWOOFER

	CHANNEL_POSITION_FRONT_LEFT_OF_CENTER  ChannelPosition = C.PA_CHANNEL_POSITION_FRONT_LEFT_OF_CENTER
	CHANNEL_POSITION_FRONT_RIGHT_OF_CENTER ChannelPosition = C.PA_CHANNEL_POSITION_FRONT_RIGHT_OF_CENTER

	CHANNEL_POSITION_SIDE_LEFT  ChannelPosition = C.PA_CHANNEL_POSITION_SIDE_LEFT
	CHANNEL_POSITION_SIDE_RIGHT ChannelPosition = C.PA_CHANNEL_POSITION_SIDE_RIGHT

	CHANNEL_POSITION_AUX0  ChannelPosition = C.PA_CHANNEL_POSITION_AUX0
	CHANNEL_POSITION_AUX1  ChannelPosition = C.PA_CHANNEL_POSITION_AUX1
	CHANNEL_POSITION_AUX2  ChannelPosition = C.PA_CHANNEL_POSITION_AUX2
	CHANNEL_POSITION_AUX3  ChannelPosition = C.PA_CHANNEL_POSITION_AUX3
	CHANNEL_POSITION_AUX4  ChannelPosition = C.PA_CHANNEL_POSITION_AUX4
	CHANNEL_POSITION_AUX5  ChannelPosition = C.PA_CHANNEL_POSITION_AUX5
	CHANNEL_POSITION_AUX6  ChannelPosition = C.PA_CHANNEL_POSITION_AUX6
	CHANNEL_POSITION_AUX7  ChannelPosition = C.PA_CHANNEL_POSITION_AUX7
	CHANNEL_POSITION_AUX8  ChannelPosition = C.PA_CHANNEL_POSITION_AUX8
	CHANNEL_POSITION_AUX9  ChannelPosition = C.PA_CHANNEL_POSITION_AUX9
	CHANNEL_POSITION_AUX10 ChannelPosition = C.PA_CHANNEL_POSITION_AUX10
	CHANNEL_POSITION_AUX11 ChannelPosition = C.PA_CHANNEL_POSITION_AUX11
	CHANNEL_POSITION_AUX12 ChannelPosition = C.PA_CHANNEL_POSITION_AUX12
	CHANNEL_POSITION_AUX13 ChannelPosition = C.PA_CHANNEL_POSITION_AUX13
	CHANNEL_POSITION_AUX14 ChannelPosition = C.PA_CHANNEL_POSITION_AUX14
	CHANNEL_POSITION_AUX15 ChannelPosition = C.PA_CHANNEL_POSITION_AUX15
	CHANNEL_POSITION_AUX16 ChannelPosition = C.PA_CHANNEL_POSITION_AUX16
	CHANNEL_POSITION_AUX17 ChannelPosition = C.PA_CHANNEL_POSITION_AUX17
	CHANNEL_POSITION_AUX18 ChannelPosition = C.PA_CHANNEL_POSITION_AUX18
	CHANNEL_POSITION_AUX19 ChannelPosition = C.PA_CHANNEL_POSITION_AUX19
	CHANNEL_POSITION_AUX20 ChannelPosition = C.PA_CHANNEL_POSITION_AUX20
	CHANNEL_POSITION_AUX21 ChannelPosition = C.PA_CHANNEL_POSITION_AUX21
	CHANNEL_POSITION_AUX22 ChannelPosition = C.PA_CHANNEL_POSITION_AUX22
	CHANNEL_POSITION_AUX23 ChannelPosition = C.PA_CHANNEL_POSITION_AUX23
	CHANNEL_POSITION_AUX24 ChannelPosition = C.PA_CHANNEL_POSITION_AUX24
	CHANNEL_POSITION_AUX25 ChannelPosition = C.PA_CHANNEL_POSITION_AUX25
	CHANNEL_POSITION_AUX26 ChannelPosition = C.PA_CHANNEL_POSITION_AUX26
	CHANNEL_POSITION_AUX27 ChannelPosition = C.PA_CHANNEL_POSITION_AUX27
	CHANNEL_POSITION_AUX28 ChannelPosition = C.PA_CHANNEL_POSITION_AUX28
	CHANNEL_POSITION_AUX29 ChannelPosition = C.PA_CHANNEL_POSITION_AUX29
	CHANNEL_POSITION_AUX30 ChannelPosition = C.PA_CHANNEL_POSITION_AUX30
	CHANNEL_POSITION_AUX31 ChannelPosition = C.PA_CHANNEL_POSITION_AUX31

	CHANNEL_POSITION_TOP_CENTER ChannelPosition = C.PA_CHANNEL_POSITION_TOP_CENTER

	CHANNEL_POSITION_TOP_FRONT_LEFT   ChannelPosition = C.PA_CHANNEL_POSITION_TOP_FRONT_LEFT
	CHANNEL_POSITION_TOP_FRONT_RIGHT  ChannelPosition = C.PA_CHANNEL_POSITION_TOP_FRONT_RIGHT
	CHANNEL_POSITION_TOP_FRONT_CENTER ChannelPosition = C.PA_CHANNEL_POSITION_TOP_FRONT_CENTER

	CHANNEL_POSITION_TOP_REAR_LEFT   ChannelPosition = C.PA_CHANNEL_POSITION_TOP_REAR_LEFT
	CHANNEL_POSITION_TOP_REAR_RIGHT  ChannelPosition = C.PA_CHANNEL_POSITION_TOP_REAR_RIGHT
	CHANNEL_POSITION_TOP_REAR_CENTER ChannelPosition = C.PA_CHANNEL_POSITION_TOP_REAR_CENTER

	CHANNEL_POSITION_MAX ChannelPosition = C.PA_CHANNEL_POSITION_MAX
)

type ChannelPositionMask uint64

// ChannelPosition.Mask makes a bitmask from a ChannelPosition.
func (p ChannelPosition) Mask() ChannelPositionMask {
	return 1 << uint(p)
}

type ChannelMapDef C.pa_channel_map_def_t

const (
	CHANNEL_MAP_AIFF    ChannelMapDef = C.PA_CHANNEL_MAP_AIFF
	CHANNEL_MAP_ALSA    ChannelMapDef = C.PA_CHANNEL_MAP_ALSA
	CHANNEL_MAP_AUX     ChannelMapDef = C.PA_CHANNEL_MAP_AUX
	CHANNEL_MAP_WAVEEX  ChannelMapDef = C.PA_CHANNEL_MAP_WAVEEX
	CHANNEL_MAP_OSS     ChannelMapDef = C.PA_CHANNEL_MAP_OSS
	CHANNEL_MAP_DEF_MAX ChannelMapDef = C.PA_CHANNEL_MAP_DEF_MAX
	CHANNEL_MAP_DEFAULT ChannelMapDef = C.PA_CHANNEL_MAP_DEFAULT
)

// ChannelMap can be used to attach labels to specific channels of a stream.
//
// These values are relevant for conversion and mixing of streams.
type ChannelMap struct {
	Channels uint8
	Map      [CHANNELS_MAX]ChannelPosition
}

func (m *ChannelMap) fromC(cmap *C.pa_channel_map) {
	m.Channels = uint8(cmap.channels)
	for i := 0; i < CHANNELS_MAX; i++ {
		m.Map[i] = ChannelPosition(cmap._map[i])
	}
}

func (m *ChannelMap) toC() *C.pa_channel_map {
	if m == nil {
		return nil
	}
	cmap := &C.pa_channel_map{}
	cmap.channels = C.uint8_t(m.Channels)
	for i := 0; i < CHANNELS_MAX; i++ {
		cmap._map[i] = C.pa_channel_position_t(m.Map[i])
	}
	return cmap
}

// ChannelMap.Init initializes the ChannelMp to a defined state,
// for which ChannelMap.Valid() will return false.
//
// calling Init() on a ChannelMap is not necessary,
// but this method is included for compatibility.
func (m *ChannelMap) Init() {
	m.fromC(C.pa_channel_map_init(m.toC()))
}

// ChannelMap.InitMono initializes the channel map for monaural audio.
func (m *ChannelMap) InitMono() {
	cmap := &C.pa_channel_map{}
	C.pa_channel_map_init_mono(cmap)
	m.fromC(cmap)
}

// ChannelMap.InitStereo initializes the channel map for stereophonic audio.
func (m *ChannelMap) InitStereo() {
	cmap := &C.pa_channel_map{}
	C.pa_channel_map_init_stereo(cmap)
	m.fromC(cmap)
}

// ChannelMap.InitAuto initializes the ChannelMap using the given mapping
// and number of channels.
//
// If there is no default channel map known for the given number of channels
// and mapping, then the ChannelMap remains unchanged and an error is returned.
func (m *ChannelMap) InitAuto(channels uint, mapping ChannelMapDef) error {
	cmap := &C.pa_channel_map{}
	mapped := C.pa_channel_map_init_auto(cmap, C.unsigned(channels), C.pa_channel_map_def_t(mapping))
	if mapped == nil {
		return fmt.Errorf("Could not map %d channels with ChannelMapDef %v", channels, mapping)
	}
	m.fromC(cmap)
	return nil
}

// ChannelMap.InitExtend initializes the ChannelMap using the given mapping
// and number of channels.
//
// If there is no default channel map known for the given number of channels
// and mapping, then it will synthesize a mapping based on a known mapping
// with fewer channels, and fill up the rest with AUX0...AUX31 channels.
func (m *ChannelMap) InitExtend(channels uint, mapping ChannelMapDef) {
	cmap := &C.pa_channel_map{}
	C.pa_channel_map_init_extend(cmap, C.unsigned(channels), C.pa_channel_map_def_t(mapping))
	m.fromC(cmap)
}

// ChannelPosition.String returns a text label for the channel position.
func (p ChannelPosition) String() string {
	cstr := C.pa_channel_position_to_string(C.pa_channel_position_t(p))
	return C.GoString(cstr)
}

// ChannelPositionFromString returns the ChannelPosition described
// by the given string.
//
// The string should be as returned by ChannelPosition.String().
func ChannelPositionFromString(s string) ChannelPosition {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))
	return ChannelPosition(C.pa_channel_position_from_string(cstr))
}

// ChannelPosition.PrettyString returns a human-readable text label
// for the channel position.
func (p ChannelPosition) PrettyString() string {
	cstr := C.pa_channel_position_to_pretty_string(C.pa_channel_position_t(p))
	return C.GoString(cstr)
}

// ChannelMap.String returns a string describing the mapping.
func (m *ChannelMap) String() string {
	s := strings.Repeat(" ", int(C.PA_CHANNEL_MAP_SNPRINT_MAX))
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))
	C.pa_channel_map_snprint(cstr, C.size_t(C.PA_CHANNEL_MAP_SNPRINT_MAX), m.toC())
	return C.GoString(cstr)
}

// ChannelMap.Parse parses a channel position list or well-known mapping name
// into a channel map structure.
//
// Input should be as returned by ChannelMap.Name() or ChannelMap.String().
func (m *ChannelMap) Parse(s string) {
	cstr := C.CString(s)
	defer C.free(unsafe.Pointer(cstr))
	m.fromC(C.pa_channel_map_parse(m.toC(), cstr))
}

// ChannelMap.Equal compares two channel maps, returning true iff they match.
func (m *ChannelMap) Equal(other *ChannelMap) bool {
	cmap1 := m.toC()
	cmap2 := other.toC()
	return C.pa_channel_map_equal(cmap1, cmap2) == 1
}

// ChannelMap.Valid returns true iff the channel map is considered valid.
func (m *ChannelMap) Valid() bool {
	cmap := m.toC()
	return C.pa_channel_map_valid(cmap) != 0
}

// ChannelMap.Compatible returns true iff compatible with the given sample spec.
func (m *ChannelMap) Compatible(spec *SampleSpec) bool {
	return C.pa_channel_map_compatible(m.toC(), spec.toC()) != 0
}

// TODO: fix terminology

// ChannelMap.Superset returns true iff every channel defined in "other"
// is also defined in the ChannelMap.
func (m *ChannelMap) Superset(other *ChannelMap) bool {
	return C.pa_channel_map_superset(m.toC(), other.toC()) != 0
}

// ChannelMap.CanBalance returns true iff applying a volume 'balance' makes sense
// with this mapping, i.e. there are left/right channels available.
func (m *ChannelMap) CanBalance() bool {
	return C.pa_channel_map_can_balance(m.toC()) != 0
}

// ChannelMap.CanFade returns true iff applying a volume 'fade' makes sense
// with this mapping, i.e. there are front/rear channels available.
func (m *ChannelMap) CanFade() bool {
	return C.pa_channel_map_can_fade(m.toC()) != 0
}

// ChannelMap.Name tries to find a well-known name for this channel mapping,
// i.e. "stereo", "surround-71" and so on.
//
// If the channel mapping is unknown, an error will be returned.
func (m *ChannelMap) Name() (string, error) {
	cstr := C.pa_channel_map_to_name(m.toC())
	if cstr == nil {
		return "", errors.New("No name found for channel mapping.")
	}
	return C.GoString(cstr), nil
}

// ChannelMap.PrettyName tries to find a human-readable text label
// for this channel mapping, i.e. "Stereo", "Surround 7.1" and so on.
//
// If the channel mapping is unknown, an error will be returned.
func (m *ChannelMap) PrettyName() (string, error) {
	cstr := C.pa_channel_map_to_pretty_name(m.toC())
	if cstr == nil {
		return "", errors.New("No name found for channel mapping.")
	}
	return C.GoString(cstr), nil
}

// ChannelMap.HasPosition returns true iff the specified channel position
// is available at least once in the channel map.
func (m *ChannelMap) HasPosition(p ChannelPosition) bool {
	return C.pa_channel_map_has_position(m.toC(), C.pa_channel_position_t(p)) != 0
}

// ChannelMap.Mask generates a bitmask from a ChannelMap.
func (m *ChannelMap) Mask() ChannelPositionMask {
	return ChannelPositionMask(C.pa_channel_map_mask(m.toC()))
}
