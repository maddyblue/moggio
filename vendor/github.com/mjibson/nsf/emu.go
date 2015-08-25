package nsf

import (
	"time"

	"github.com/mjibson/nsf/cpu6502"
)

const (
	// 1.79 MHz
	cpuClock = 236250000 / 11 / 12
)

var (
	// DefaultSampleRate is the default sample rate of a track after calling
	// Init().
	DefaultSampleRate int64 = 44100

	DefaultDuration = time.Minute * 2
	DefaultFade     = time.Second * 2
	DefaultSilence  = time.Second * 2
)

type Song struct {
	Name string
	// Duration is the duration after which Play will halt. Set to < 0 to play
	// indefinitely.
	Duration time.Duration
	// After Duration, fade out. Set to 0 to end immediately.
	Fade time.Duration
}

type NSF struct {
	*cpu6502.Cpu

	// Silence is the duration for which if the result of Play is silence,
	// Play will halt. Set to 0 to disable silence check.
	Silence time.Duration
	// SampleRate is the sample rate at which samples will be generated. If not
	// set before Init(), it is set to DefaultSampleRate.
	SampleRate int64

	// Start is the 0-based index of the starting song
	Start     byte
	Songs     []Song
	Copyright string
	Artist    string
	Game      string

	LoadAddr uint16
	InitAddr uint16
	PlayAddr uint16

	SpeedNTSC  uint16
	Bankswitch [8]byte
	Data       []byte

	ram         *ram
	totalTicks  int64
	frameTicks  int64
	sampleTicks int64
	playTicks   int64
	samples     []float32
	prevs       [4]float32
	pi          int // prevs index

	silent time.Duration
	played time.Duration
	zero   bool
	// song is the currently playing song.
	song Song
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

// Init initializes the 1-based song for playing. Only one song my play
// at once. An invalid song index will play the first song.
func (n *NSF) Init(song int) {
	if len(n.Songs) < song || song < 0 {
		song = 1
	}
	n.song = n.Songs[song-1]
	if n.SampleRate == 0 {
		n.SampleRate = DefaultSampleRate
	}
	n.ram = new(ram)
	copy(n.ram.M[n.LoadAddr:], n.Data)
	n.Cpu = cpu6502.New(n.ram)
	n.Cpu.DisableDecimal = true
	n.Cpu.P = 0x24
	n.Cpu.S = 0xfd
	n.ram.A.Init()
	n.Cpu.A = byte(song - 1)
	n.Cpu.PC = n.InitAddr
	n.Cpu.Run()
	n.Cpu.T = n
}

func (n *NSF) step() {
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
	if n.song.Duration > 0 && n.played > n.song.Duration {
		return nil
	}
	ticksPerPlay := int64(playDur / (time.Second / cpuClock))
	n.samples = make([]float32, 0, samples)
	n.zero = true
	for len(n.samples) < samples {
		n.playTicks = 0
		n.Cpu.PC = n.PlayAddr
		for n.Cpu.PC != 0 && len(n.samples) < samples {
			n.step()
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
