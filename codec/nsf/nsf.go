package nsf

import (
	"errors"
	"io"
	"io/ioutil"
)

var (
	ErrUnrecognized = errors.New("nsf: unrecognized format")
)

const (
	NSF_HEADER_LEN = 0x80
	NSF_VERSION    = 0x5
	NSF_SONGS      = 0x6
	NSF_START      = 0x7
	NSF_LOAD       = 0x8
	NSF_INIT       = 0xa
	NSF_PLAY       = 0xc
	NSF_SONG       = 0xe
	NSF_ARTIST     = 0x2e
	NSF_COPYRIGHT  = 0x4e
	NSF_SPEED_NTSC = 0x6e
	NSF_BANKSWITCH = 0x70
	NSF_SPEED_PAL  = 0x78
	NSF_PAL_NTSC   = 0x7a
	NSF_EXTRA      = 0x7b
	NSF_ZERO       = 0x7c
)

func ReadNSF(r io.Reader) (n *NSF, err error) {
	n = &NSF{}
	n.b, err = ioutil.ReadAll(r)
	if err != nil {
		return
	}
	if len(n.b) < NSF_HEADER_LEN ||
		string(n.b[0:NSF_VERSION]) != "NESM\u001a" {
		return nil, ErrUnrecognized
	}
	n.Version = n.b[NSF_VERSION]
	n.Songs = n.b[NSF_SONGS]
	n.Start = n.b[NSF_START]
	n.Load = bLEtoUint16(n.b[NSF_LOAD:])
	n.Init = bLEtoUint16(n.b[NSF_INIT:])
	n.Play = bLEtoUint16(n.b[NSF_PLAY:])
	n.Song = bToString(n.b[NSF_SONG:])
	n.Artist = bToString(n.b[NSF_ARTIST:])
	n.Copyright = bToString(n.b[NSF_COPYRIGHT:])
	n.SpeedNTSC = bLEtoUint16(n.b[NSF_SPEED_NTSC:])
	copy(n.Bankswitch[:], n.b[NSF_BANKSWITCH:NSF_SPEED_PAL])
	n.SpeedPAL = bLEtoUint16(n.b[NSF_SPEED_PAL:])
	n.PALNTSC = n.b[NSF_PAL_NTSC]
	n.Extra = n.b[NSF_EXTRA]
	return
}

type NSF struct {
	b []byte

	Version byte
	Songs   byte
	Start   byte

	Load uint16
	Init uint16
	Play uint16

	Song      string
	Artist    string
	Copyright string

	SpeedNTSC  uint16
	Bankswitch [8]byte
	SpeedPAL   uint16
	PALNTSC    byte
	Extra      byte
	Data       []byte
}

// little-endian [2]byte to uint16 conversion
func bLEtoUint16(b []byte) uint16 {
	return uint16(b[1])<<8 + uint16(b[0])
}

// null-terminated bytes to string
func bToString(b []byte) string {
	i := 0
	for i = range b {
		if b[i] == 0 {
			break
		}
	}
	return string(b[:i])
}
