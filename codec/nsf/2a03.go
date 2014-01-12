package nsf

type Apu struct {
	S1, S2 Square

	Odd        bool
	FC         byte
	FT         byte
	IrqDisable bool
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
	Negate    bool
	Period    byte
	Enable    bool
	Divider   byte
	Reset     bool
	NegOffset int
}

type Envelope struct {
	Volume   byte
	Divider  byte
	Counter  byte
	Loop     bool
	Constant bool
	Start    bool
}

type Timer struct {
	Tick   uint16
	Length uint16
}

type Length struct {
	Halt    bool
	Counter byte
}

func (a *Apu) Init() {
	a.S1.Sweep.NegOffset = -1
	for i := uint16(0x4000); i <= 0x400f; i++ {
		a.Write(i, 0)
	}
	a.Write(0x4010, 0x10)
	a.Write(0x4011, 0)
	a.Write(0x4012, 0)
	a.Write(0x4013, 0)
	a.Write(0x4015, 0xf)
	a.Write(0x4017, 0)
}

func (a *Apu) Write(v uint16, b byte) {
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
		a.FT = 0
		if b&0x80 != 0 {
			a.FC = 5
			a.FrameStep()
		} else {
			a.FC = 4
		}
		a.IrqDisable = b&0x40 != 0
	}
}

func (s *Square) Control1(b byte) {
	s.Envelope.Control(b)
	s.Duty.Control(b)
	s.Length.Halt = b&0x20 != 0
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

	s.Envelope.Start = true
	s.Duty.Counter = 0
}

func (d *Duty) Control(b byte) {
	d.Type = b >> 6
}

func (s *Sweep) Control(b byte) {
	s.Shift = b & 0x7
	s.Negate = b&0x8 != 0
	s.Period = (b >> 4) & 0x7
	s.Enable = b&0x80 != 0
	s.Reset = true
}

func (e *Envelope) Control(b byte) {
	e.Volume = b & 0xf
	e.Constant = b&0x10 != 0
	e.Loop = b&0x20 != 0
}

func (l *Length) Set(b byte) {
	if !l.Halt {
		l.Counter = LenLookup[b]
	}
}

func (l *Length) Enabled() bool {
	return l.Counter != 0
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
	return b
}

func (d *Duty) Clock() {
	if d.Counter == 0 {
		d.Counter = 7
	} else {
		d.Counter--
	}
}

func (s *Sweep) Clock() (r bool) {
	if s.Divider == 0 {
		s.Divider = s.Period
		r = true
	} else {
		s.Divider--
	}
	if s.Reset {
		s.Divider = 0
		s.Reset = false
	}
	return
}

func (e *Envelope) Clock() {
	if e.Start {
		e.Start = false
		e.Counter = 15
	} else {
		if e.Divider == 0 {
			e.Divider = e.Volume
			if e.Counter != 0 {
				e.Counter--
			} else if e.Loop {
				e.Counter = 15
			}
		} else {
			e.Divider--
		}
	}
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
	return t.Tick == t.Length
}

func (s *Square) Clock() {
	if s.Timer.Clock() {
		s.Duty.Clock()
	}
}

func (a *Apu) Step() {
	if a.Odd {
		if a.S1.Enable {
			a.S1.Clock()
		}
		if a.S2.Enable {
			a.S2.Clock()
		}
	}
	a.Odd = !a.Odd
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
		a.S1.FrameStep()
		a.S2.FrameStep()
	}
	if a.FC == 4 && a.FT == 3 && !a.IrqDisable {
		// todo: assert cpu irq line
	}
}

func (s *Square) FrameStep() {
	s.Length.Clock()
	if s.Sweep.Clock() && s.Sweep.Enable && s.Sweep.Shift > 0 {
		r := s.SweepResult()
		if r <= 0x7ff {
			s.Timer.Tick = r
		}
	}
}

func (l *Length) Clock() {
	if !l.Halt && l.Counter > 0 {
		l.Counter--
	}
}

func (a *Apu) Volume() float32 {
	p := PulseOut[a.S1.Volume()+a.S2.Volume()]
	return p
}

func (s *Square) Volume() uint8 {
	if s.Enable && s.Duty.Enabled() && s.Length.Enabled() && s.Timer.Tick >= 8 && s.SweepResult() <= 0x7ff {
		return s.Envelope.Output()
	}
	return 0
}

func (e *Envelope) Output() byte {
	if e.Constant {
		return e.Volume
	}
	return e.Counter
}

func (s *Square) SweepResult() uint16 {
	r := int(s.Timer.Tick >> s.Sweep.Shift)
	if s.Sweep.Negate {
		r =  - r
	}
	r += int(s.Timer.Tick)
	if r > 0x7ff {
		r = 0x800
	}
	return uint16(r)
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
		0x0a, 0xfe, 0x14, 0x02,
		0x28, 0x04, 0x50, 0x06,
		0xa0, 0x08, 0x3c, 0x0a,
		0x0e, 0x0c, 0x1a, 0x0e,
		0x0c, 0x10, 0x18, 0x12,
		0x30, 0x14, 0x60, 0x16,
		0xc0, 0x18, 0x48, 0x1a,
		0x10, 0x1c, 0x20, 0x1e,
	}
)

func init() {
	for i := range PulseOut {
		PulseOut[i] = 95.88 / (8128/float32(i) + 100)
	}
}
