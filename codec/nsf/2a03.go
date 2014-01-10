package nsf

import "fmt"

type Apu struct {
	S1, S2 Square

	FC        byte
	FT        byte
	IrqEnable bool
}

type Square struct {
	Envelope
	Timer
	Length
	Sweep
	Duty

	Enable bool
}

type Duty struct {
	Type    byte
	Counter byte
}

type Sweep struct {
	Shift     byte
	Decrease  bool
	Rate      byte
	Enable    bool
	Counter   byte
	Value     uint16
	NegOffset uint16
}

type Envelope struct {
	DecayRate    byte
	DecayCounter byte
	DecayDisable bool
	Loop         bool
	Counter      byte
}

type Timer struct {
	Tick   uint16
	Length uint16
}

type Length struct {
	Disable bool
	Counter byte
}

func (a *Apu) Write(v uint16, b byte) {
	fmt.Printf("WRITE %x %08b\n", v, b)
	switch v & 0xff {
	case 0x00:
		a.S1.Control1(b)
	case 0x01:
		a.S1.Control2(b)
	case 0x02:
		a.S1.Control3(b)
	case 0x03:
		a.S1.Control4(b)
	case 0x04:
		a.S2.Control1(b)
	case 0x05:
		a.S2.Control2(b)
	case 0x06:
		a.S2.Control3(b)
	case 0x07:
		a.S2.Control4(b)
	case 0x15:
		a.S1.Disable(b&0x1 == 0)
		a.S2.Disable(b&0x2 == 0)
	case 0x17:
		if b&0x80 != 0 {
			a.FC = 5
		} else {
			a.FC = 4
		}
		a.IrqEnable = b&0x40 != 0
	}
}

func (s *Square) Control1(b byte) {
	s.Envelope.Control(b)
	s.Duty.Control(b)
	s.Length.Disable = b&0x20 != 0
}

func (s *Square) Control2(b byte) {
	s.Sweep.Control(b)
}

func (s *Square) Control3(b byte) {
	s.Timer.Length &= 0xff00
	s.Timer.Length |= uint16(b)
}

func (s *Square) Control4(b byte) {
	s.Timer.Length &= 0xff
	s.Timer.Length |= uint16(b&0x7) << 8
	s.Length.Set(b >> 3)

	s.Envelope.Counter = 15
	s.Duty.Counter = 0
}

func (d *Duty) Control(b byte) {
	d.Type = b >> 6
}

func (s *Sweep) Control(b byte) {
	s.Shift = b & 0x7
	s.Decrease = b&0x8 != 0
	s.Rate = (b >> 4) & 0x7
	s.Enable = b&0x80 != 0
}

func (e *Envelope) Control(b byte) {
	e.DecayRate = b & 0xf
	e.DecayDisable = b&0x10 != 0
	e.Loop = b&0x20 != 0

	fmt.Printf("ENV CONTROL %08b, %v, %v, %v\n", b, e.DecayRate, e.DecayDisable, e.Loop)
	e.Counter = 15
}

func (l *Length) Set(b byte) {
	l.Counter = LenLookup[b]
}

func (l *Length) Enabled() bool {
	return l.Counter != 0 && !l.Disable
}

func (s *Square) Disable(b bool) {
	s.Enable = !b
	if b {
		s.Length.Counter = 0
	}
}

func (a *Apu) Read(v uint16) byte {
	var b byte
	if v == 0x4015 {
		if a.S1.Length.Counter > 0 {
			b |= 0x1
		}
		if a.S2.Length.Counter > 0 {
			b |= 0x2
		}
	}
	fmt.Printf("READ %x %08b\n", v, b)
	return b
}

func (d *Duty) Clock() {
	if d.Counter == 0 {
		d.Counter = 7
	} else {
		d.Counter--
	}
	//println("DUTY", d.Counter)
}

func (s *Sweep) Clock() {
	if s.Counter == 0 {
		s.Counter = s.Rate
	} else {
		s.Counter--
		// todo: only update if chan's len counter != 0
		if s.Shift > 0 && s.Enable && s.Value >= 8 {
			d := s.Value >> s.Shift
			if s.Decrease {
				d = s.NegOffset - d
			}
			if s.Value+d < 0x800 {
				s.Value += d
			}
		}
	}
}

func (e *Envelope) Clock() {
	if e.DecayCounter == 0 {
		e.DecayCounter = e.DecayRate + 1
		if e.Counter > 0 {
			e.Counter--
		} else if e.Loop {
			e.Counter = 15
		}
	} else {
		e.DecayCounter--
	}
	//println("ENVL", e.Counter, e.DecayCounter)
}

// 1.79 MHz/(N+1)
// square: -> duty cycle generator
// triangle: -> triangle step generator
// noise: -> random number generator
func (t *Timer) Clock() bool {
	if t.Tick == 0 {
		t.Tick = t.Length
	} else {
		t.Tick--
	}
	//println("TIMR", t.Tick, t.Length)
	return t.Tick == t.Length
}

func (s *Square) Clock() {
	if s.Timer.Clock() {
		s.Duty.Clock()
	}
}

func (a *Apu) Step() {
	if a.S1.Enable {
		a.S1.Clock()
	}
	if a.S2.Enable {
		a.S2.Clock()
	}
}

func (a *Apu) FrameStep() {
	a.FT++
	if a.FT == a.FC {
		a.FT = 0
	}
	if a.FT <= 3 {
		a.S1.Envelope.Clock()
	}
	if a.FT == 1 || a.FT == 3 {
		a.S1.Length.Clock()
		a.S1.Sweep.Clock()
	}
	println("FRAM", a.FT, a.FC)
}

func (l *Length) Clock() {
	if !l.Disable && l.Counter > 0 {
		l.Counter--
	}
	//println("LENG", l.Counter)
}

func (a *Apu) Volume() float32 {
	p := PulseOut[a.S1.Volume()+a.S2.Volume()]
	return p
}

func (s *Square) Volume() uint8 {
	if s.Enable && s.Duty.Enabled() && s.Length.Enabled() && s.Timer.Tick >= 8 {
		return s.Envelope.Volume()
	}
	return 0
}

func (e *Envelope) Volume() byte {
	if e.DecayDisable {
		return e.DecayRate
	}
	return e.Counter
}

func (d *Duty) Enabled() bool {
	return DutyCycle[d.Type][d.Counter] == 1
}

var (
	PulseOut  [32]float32
	DutyCycle = [4][8]byte{
		{0, 1, 0, 0, 0, 0, 0, 0},
		{0, 1, 1, 0, 0, 0, 0, 0},
		{0, 1, 1, 1, 1, 0, 0, 0},
		{1, 0, 0, 1, 1, 1, 1, 1},
	}
	LenLookup = []byte{
		0x05, 0x0a, 0x14, 0x28,
		0x50, 0x1e, 0x07, 0x0d,
		0x06, 0x0c, 0x18, 0x30,
		0x60, 0x24, 0x08, 0x10,
		0x7f, 0x01, 0x02, 0x03,
		0x04, 0x05, 0x06, 0x07,
		0x08, 0x09, 0x0a, 0x0b,
		0x0c, 0x0d, 0x0e, 0x0f,
	}
)

func init() {
	for i := range PulseOut {
		PulseOut[i] = 95.88 / (8128/float32(i) + 100)
	}
}
