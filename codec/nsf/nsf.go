package nsf

import (
	"errors"
	"io"
	"io/ioutil"
	"time"

	"github.com/mjibson/mog/codec/nsf/cpu6502"
)

const (
	// 1.79 MHz
	cpuClock          = 236250000 / 11 / 12
	DefaultSampleRate = 44100
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
	n.LoadAddr = bLEtoUint16(n.b[NSF_LOAD:])
	n.InitAddr = bLEtoUint16(n.b[NSF_INIT:])
	n.PlayAddr = bLEtoUint16(n.b[NSF_PLAY:])
	n.Song = bToString(n.b[NSF_SONG:])
	n.Artist = bToString(n.b[NSF_ARTIST:])
	n.Copyright = bToString(n.b[NSF_COPYRIGHT:])
	n.SpeedNTSC = bLEtoUint16(n.b[NSF_SPEED_NTSC:])
	copy(n.Bankswitch[:], n.b[NSF_BANKSWITCH:NSF_SPEED_PAL])
	n.SpeedPAL = bLEtoUint16(n.b[NSF_SPEED_PAL:])
	n.PALNTSC = n.b[NSF_PAL_NTSC]
	n.Extra = n.b[NSF_EXTRA]
	n.Data = n.b[NSF_HEADER_LEN:]
	return
}

type NSF struct {
	*Ram
	*cpu6502.Cpu

	b []byte // raw NSF data

	Version byte
	Songs   byte
	Start   byte

	LoadAddr uint16
	InitAddr uint16
	PlayAddr uint16

	Song      string
	Artist    string
	Copyright string

	SpeedNTSC  uint16
	Bankswitch [8]byte
	SpeedPAL   uint16
	PALNTSC    byte
	Extra      byte
	Data       []byte

	SampleRate  int64
	totalTicks  int64
	frameTicks  int64
	sampleTicks int64
	playTicks   int64
	samples     []float32
	prevs       [4]float32
	pi          int // prevs index
}

func (n *NSF) Tick() {
	n.Ram.A.Step()
	n.totalTicks++
	n.frameTicks++
	if n.frameTicks == cpuClock/240 {
		n.frameTicks = 0
		n.Ram.A.FrameStep()
	}
	n.sampleTicks++
	if n.sampleTicks >= cpuClock/n.SampleRate {
		n.sampleTicks = 0
		n.append(n.Ram.A.Volume())
	}
	n.playTicks++
}

func (n *NSF) append(v float32) {
	n.prevs[n.pi] = v
	n.pi++
	if n.pi >= len(n.prevs) {
		n.pi = 0
	}
	var sum float32
	for _, s := range n.prevs {
		sum += s
	}
	sum /= float32(len(n.prevs))
	n.samples = append(n.samples, sum)
}

func (n *NSF) Init(song byte) {
	if n.SampleRate == 0 {
		n.SampleRate = DefaultSampleRate
	}
	n.Ram = new(Ram)
	n.Cpu = cpu6502.New(n.Ram)
	copy(n.Ram.M[n.LoadAddr:], n.Data)
	n.Ram.A.Init()
	n.Cpu.A = song - 1
	n.Cpu.PC = n.InitAddr
	n.Cpu.T = nil
	n.Cpu.Run()
	n.Cpu.T = n
}

func (n *NSF) Play(samples int) []float32 {
	playDur := time.Duration(n.SpeedNTSC) * time.Nanosecond * 1000
	ticksPerPlay := int64(playDur / (time.Second / cpuClock))
	n.samples = make([]float32, 0, samples)
	for len(n.samples) < samples {
		n.playTicks = 0
		n.Cpu.PC = n.PlayAddr
		n.Cpu.Halt = false
		for !n.Cpu.Halt && len(n.samples) < samples {
			n.Cpu.Step()
			if !n.Cpu.I() {
				panic("INTERRUPT")
			}
		}
		for i := ticksPerPlay - n.playTicks; i > 0 && len(n.samples) < samples; i-- {
			n.Tick()
		}
	}
	return n.samples
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

type Ram struct {
	M [0xffff + 1]byte
	A Apu
}

func (r *Ram) Read(v uint16) byte {
	if v == 0x4015 {
		return r.A.Read(v)
	} else {
		return r.M[v]
	}
}

func (r *Ram) Write(v uint16, b byte) {
	r.M[v] = b
	if v&0xf000 == 0x4000 {
		r.A.Write(v, b)
	}
}
