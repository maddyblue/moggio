package nsf

import (
	"errors"
	"io"
	"io/ioutil"
	"time"

	"github.com/mjibson/mog/_third_party/github.com/mjibson/nsf/cpu6502"
)

const (
	// 1.79 MHz
	cpuClock = 236250000 / 11 / 12
)

var (
	// DefaultSampleRate is the default sample rate of a track after calling
	// Init().
	DefaultSampleRate int64 = 44100
	ErrUnrecognized         = errors.New("nsf: unrecognized format")
)

const (
	nsfHEADER_LEN = 0x80
	nsfVERSION    = 0x5
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
	nsfPAL_NTSC   = 0x7a
	nsfEXTRA      = 0x7b
	nsfZERO       = 0x7c
)

func ReadNSF(r io.Reader) (n *NSF, err error) {
	n = New()
	n.b, err = ioutil.ReadAll(r)
	if err != nil {
		return
	}
	if len(n.b) < nsfHEADER_LEN ||
		string(n.b[0:nsfVERSION]) != "NESM\u001a" {
		return nil, ErrUnrecognized
	}
	n.Version = n.b[nsfVERSION]
	n.Songs = n.b[nsfSONGS]
	n.Start = n.b[nsfSTART]
	n.LoadAddr = bLEtoUint16(n.b[nsfLOAD:])
	n.InitAddr = bLEtoUint16(n.b[nsfINIT:])
	n.PlayAddr = bLEtoUint16(n.b[nsfPLAY:])
	n.Song = bToString(n.b[nsfSONG:])
	n.Artist = bToString(n.b[nsfARTIST:])
	n.Copyright = bToString(n.b[nsfCOPYRIGHT:])
	n.SpeedNTSC = bLEtoUint16(n.b[nsfSPEED_NTSC:])
	copy(n.Bankswitch[:], n.b[nsfBANKSWITCH:nsfSPEED_PAL])
	n.SpeedPAL = bLEtoUint16(n.b[nsfSPEED_PAL:])
	n.PALNTSC = n.b[nsfPAL_NTSC]
	n.Extra = n.b[nsfEXTRA]
	n.Data = n.b[nsfHEADER_LEN:]
	if n.SampleRate == 0 {
		n.SampleRate = DefaultSampleRate
	}
	copy(n.ram.M[n.LoadAddr:], n.Data)
	return
}

type NSF struct {
	*ram
	*cpu6502.Cpu

	b []byte // raw NSF data

	// Silence is the duration for which if the result of Play is silence,
	// Play will halt. Defaults to 1s. Set to 0 to disable silence check.
	Silence time.Duration
	// Limit is the duration after which Play will halt. Defaults to 2m. Set to
	// 0 to play indefinitely.
	Limit time.Duration

	silent time.Duration
	played time.Duration
	zero   bool

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

	// SampleRate is the sample rate at which samples will be generated. If not
	// set before Init(), it is set to DefaultSampleRate.
	SampleRate  int64
	totalTicks  int64
	frameTicks  int64
	sampleTicks int64
	playTicks   int64
	samples     []float32
	prevs       [4]float32
	pi          int // prevs index
}

func New() *NSF {
	n := NSF{
		Silence: time.Second,
		Limit:   time.Minute * 2,

		ram: new(ram),
	}
	n.Cpu = cpu6502.New(n.ram)
	n.Cpu.T = &n
	n.Cpu.DisableDecimal = true
	n.Cpu.P = 0x24
	n.Cpu.S = 0xfd
	return &n
}

func (n *NSF) Tick() {
	n.ram.A.Step()
	n.totalTicks++
	n.frameTicks++
	if n.frameTicks == cpuClock/240 {
		n.frameTicks = 0
		n.ram.A.FrameStep()
	}
	n.sampleTicks++
	if n.SampleRate > 0 && n.sampleTicks >= cpuClock/n.SampleRate {
		n.sampleTicks = 0
		n.append(n.ram.A.Volume())
	}
	n.playTicks++
}

func (n *NSF) append(v float32) {
	if v != 0 {
		n.zero = false
	}
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

func (n *NSF) Init(song int) {
	n.ram.A.Init()
	n.Cpu.A = byte(song - 1)
	n.Cpu.PC = n.InitAddr
	n.Cpu.T = nil
	n.Cpu.Run()
	n.Cpu.T = n
}

func (n *NSF) Step() {
	n.Cpu.Step()
	if !n.Cpu.I() && n.ram.A.Interrupt {
		n.Cpu.Interrupt()
	}
}

// Play returns the requested number of samples. If less are returned,
// the silence check or time limit have been reached.
func (n *NSF) Play(samples int) []float32 {
	playDur := time.Duration(n.SpeedNTSC) * time.Nanosecond * 1000
	sampleDur := time.Duration(samples) * time.Second / time.Duration(n.SampleRate)
	n.played += sampleDur
	if n.Limit > 0 && n.played > n.Limit {
		return nil
	}
	ticksPerPlay := int64(playDur / (time.Second / cpuClock))
	n.samples = make([]float32, 0, samples)
	n.zero = true
	for len(n.samples) < samples {
		n.playTicks = 0
		n.Cpu.PC = n.PlayAddr
		for n.Cpu.PC != 0 && len(n.samples) < samples {
			n.Step()
		}
		for i := ticksPerPlay - n.playTicks; i > 0 && len(n.samples) < samples; i-- {
			n.Tick()
		}
	}
	if n.zero {
		n.silent += sampleDur
		if n.Silence > 0 && n.silent > n.Silence {
			return nil
		}
	} else {
		n.silent = 0
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

type ram struct {
	M [0xffff + 1]byte
	A apu
}

func (r *ram) Read(v uint16) byte {
	switch v {
	case 0x4015:
		return r.A.Read(v)
	default:
		return r.M[v]
	}
}

func (r *ram) Write(v uint16, b byte) {
	r.M[v] = b
	if v&0xf000 == 0x4000 {
		r.A.Write(v, b)
	}
}

func (n *NSF) Seek(t time.Time) {
	// todo: implement
}
