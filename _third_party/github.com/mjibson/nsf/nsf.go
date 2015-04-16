package nsf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"
)

var ErrUnrecognized = errors.New("nsf: unrecognized format")

const (
	nsfHEADER_LEN = 0x80
	nsfSONGS      = 0x6
	nsfSTART      = 0x7
	nsfLOAD       = 0x8
	nsfINIT       = 0xa
	nsfPLAY       = 0xc
	nsfSONG       = 0xe
	nsfARTIST     = 0x2e
	nsfCOPYRIGHT  = 0x4e
	nsfSPEED_NTSC = 0x6e
	nsfBANKSWITCH = 0x70
	nsfSPEED_PAL  = 0x78
)

func New(r io.Reader) (*NSF, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	n, err := ReadNSF(b)
	if err == nil {
		return n, nil
	}
	return ReadNSFE(b)
}

// ReadNSF reads a NSF file from b.
func ReadNSF(b []byte) (*NSF, error) {
	if len(b) < nsfHEADER_LEN || !bytes.HasPrefix(b, []byte("NESM\u001a")) {
		return nil, ErrUnrecognized
	}
	var n NSF
	n.Songs = make([]Song, int(b[nsfSONGS]))
	for i := range n.Songs {
		n.Songs[i] = Song{
			Duration: DefaultDuration,
		}
	}
	n.Start = b[nsfSTART]
	n.LoadAddr = bLEtoUint16(b[nsfLOAD:])
	n.InitAddr = bLEtoUint16(b[nsfINIT:])
	n.PlayAddr = bLEtoUint16(b[nsfPLAY:])
	n.Game = bToString(b[nsfSONG:])
	n.Artist = bToString(b[nsfARTIST:])
	n.Copyright = bToString(b[nsfCOPYRIGHT:])
	n.SpeedNTSC = bLEtoUint16(b[nsfSPEED_NTSC:])
	copy(n.Bankswitch[:], b[nsfBANKSWITCH:nsfSPEED_PAL])
	n.Data = b[nsfHEADER_LEN:]
	return &n, nil
}

// ReadNSFE reads a NSFE file from b.
func ReadNSFE(b []byte) (*NSF, error) {
	if !bytes.HasPrefix(b, []byte("NSFE")) {
		return nil, ErrUnrecognized
	}
	var n NSF
	n.SpeedNTSC = 16666
	b = b[4:]
	for {
		if len(b) < 8 {
			return nil, ErrUnrecognized
		}
		size := binary.LittleEndian.Uint32(b)
		id := string(b[4:8])
		if id != "INFO" && n.Songs == nil {
			return nil, fmt.Errorf("nsf: INFO chunk not first")
		}
		if id == "NEND" {
			break
		}
		b = b[8:]
		if uint32(len(b)) < size {
			return nil, ErrUnrecognized
		}
		data := b[:size]
		b = b[size:]
		switch id {
		case "INFO":
			n.LoadAddr = bLEtoUint16(data)
			n.InitAddr = bLEtoUint16(data[2:])
			n.PlayAddr = bLEtoUint16(data[4:])
			if data[7] != 0 {
				return nil, fmt.Errorf("nsf: unsupported sound chip: %02x", data[7])
			}
			n.Songs = make([]Song, data[8])
			n.Start = data[9]
		case "DATA":
			n.Data = data
		case "BANK":
			copy(n.Bankswitch[:], data)
		case "time":
			for i := 0; len(data) > 4; data, i = data[4:], i+1 {
				tm := int32(binary.LittleEndian.Uint32(data))
				ms := time.Duration(tm) * time.Millisecond
				n.Songs[i].Duration = ms
			}
		case "fade":
			for i := 0; len(data) > 4; data, i = data[4:], i+1 {
				tm := int32(binary.LittleEndian.Uint32(data))
				ms := time.Duration(tm) * time.Millisecond
				n.Songs[i].Fade = ms
			}
		case "auth":
			ss := nullStrings(data)
			if len(ss) != 4 {
				return nil, fmt.Errorf("nsf: bad auth chunk")
			}
			n.Game = ss[0]
			n.Artist = ss[1]
			n.Copyright = ss[2]
		case "tlbl":
			for i, s := range nullStrings(data) {
				if i >= len(n.Songs) {
					break
				}
				n.Songs[i].Name = s
			}
		case "plst", "text":
			break
		default:
			panic(id)
		}
	}
	return &n, nil
}

func nullStrings(b []byte) []string {
	return strings.FieldsFunc(string(b), func(r rune) bool {
		return r == 0
	})
}
