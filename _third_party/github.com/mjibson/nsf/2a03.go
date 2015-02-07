package nsf

type apu struct {
	S1, S2 square
	triangle
	noise

	Odd        bool
	FC         byte
	FT         byte
	IrqDisable bool
	Interrupt  bool
}

type noise struct {
	envelope
	timer
	length
	Short bool
	Shift uint16

	Enable bool
}

type triangle struct {
	linear
	timer
	length
	SI int // sequence index

	Enable bool
}

type linear struct {
	Reload  byte
	Halt    bool
	Flag    bool
	Counter byte
}

type square struct {
	envelope
	timer
	length
	sweep
	duty

	Enable bool
}

type duty struct {
	Type    byte
	Counter byte
}

type sweep struct {
	Shift     byte
	Negate    bool
	Period    byte
	Enable    bool
	Divider   byte
	Reset     bool
	NegOffset int
}

type envelope struct {
	Volume   byte
	Divider  byte
	Counter  byte
	Loop     bool
	Constant bool
	Start    bool
}

type timer struct {
	Tick   uint16
	length uint16
}

type length struct {
	Halt    bool
	Counter byte
}

func (a *apu) Init() {
	a.S1.sweep.NegOffset = -1
	for i := uint16(0x4000); i <= 0x400f; i++ {
		a.Write(i, 0)
	}
	a.Write(0x4010, 0x10)
	a.Write(0x4011, 0)
	a.Write(0x4012, 0)
	a.Write(0x4013, 0)
	a.Write(0x4015, 0xf)
	a.Write(0x4017, 0)
	a.noise.Shift = 1
}

func (a *apu) Write(v uint16, b byte) {
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
	case 0x08:
		a.triangle.Control1(b)
	case 0x0a:
		a.triangle.Control2(b)
	case 0x0b:
		a.triangle.Control3(b)
	case 0x0c:
		a.noise.Control1(b)
	case 0x0e:
		a.noise.Control2(b)
	case 0x0f:
		a.noise.Control3(b)
	case 0x15:
		a.S1.Disable(b&0x1 == 0)
		a.S2.Disable(b&0x2 == 0)
		a.triangle.Disable(b&0x4 == 0)
		a.noise.Disable(b&0x8 == 0)
	case 0x17:
		a.FT = 0
		if b&0x80 != 0 {
			a.FC = 5
			a.FrameStep()
		} else {
			a.FC = 4
		}
		a.IrqDisable = b&0x40 != 0
		if a.IrqDisable && a.Interrupt {
			a.Interrupt = false
		}
	}
}

func (n *noise) Control1(b byte) {
	n.envelope.Control(b)
}

func (n *noise) Control2(b byte) {
	n.timer.length = noiseLookup[b&0xf]
	n.Short = b&0x8 != 0
}

func (n *noise) Control3(b byte) {
	n.length.Set(b >> 3)
}

func (t *triangle) Control1(b byte) {
	t.linear.Control(b)
	t.length.Halt = b&0x80 != 0
}

func (l *linear) Control(b byte) {
	l.Flag = b&0x80 != 0
	l.Reload = b & 0x7f
}

func (t *triangle) Control2(b byte) {
	t.timer.length &= 0xff00
	t.timer.length |= uint16(b)
}

func (t *triangle) Control3(b byte) {
	t.timer.length &= 0xff
	t.timer.length |= uint16(b&0x7) << 8
	t.length.Set(b >> 3)
	t.linear.Halt = true
}

func (s *square) Control1(b byte) {
	s.envelope.Control(b)
	s.duty.Control(b)
	s.length.Halt = b&0x20 != 0
}

func (s *square) Control2(b byte) {
	s.sweep.Control(b)
}

func (s *square) Control3(b byte) {
	s.timer.length &= 0xff00
	s.timer.length |= uint16(b)
}

func (s *square) Control4(b byte) {
	s.timer.length &= 0xff
	s.timer.length |= uint16(b&0x7) << 8
	s.length.Set(b >> 3)

	s.envelope.Start = true
	s.duty.Counter = 0
}

func (d *duty) Control(b byte) {
	d.Type = b >> 6
}

func (s *sweep) Control(b byte) {
	s.Shift = b & 0x7
	s.Negate = b&0x8 != 0
	s.Period = (b >> 4) & 0x7
	s.Enable = b&0x80 != 0
	s.Reset = true
}

func (e *envelope) Control(b byte) {
	e.Volume = b & 0xf
	e.Constant = b&0x10 != 0
	e.Loop = b&0x20 != 0
}

func (l *length) Set(b byte) {
	l.Counter = lenLookup[b]
}

func (l *length) Enabled() bool {
	return l.Counter != 0
}

func (s *square) Disable(b bool) {
	s.Enable = !b
	if b {
		s.length.Counter = 0
	}
}

func (t *triangle) Disable(b bool) {
	t.Enable = !b
	if b {
		t.length.Counter = 0
	}
}

func (n *noise) Disable(b bool) {
	n.Enable = !b
	if b {
		n.length.Counter = 0
	}
}

func (a *apu) Read(v uint16) byte {
	var b byte
	if v == 0x4015 {
		if a.S1.length.Counter > 0 {
			b |= 0x1
		}
		if a.S2.length.Counter > 0 {
			b |= 0x2
		}
		if a.triangle.length.Counter > 0 {
			b |= 0x4
		}
		if a.noise.length.Counter > 0 {
			b |= 0x8
		}
		if a.Interrupt {
			b |= 0x40
			a.Interrupt = false
		}
	}
	return b
}

func (d *duty) Clock() {
	if d.Counter == 0 {
		d.Counter = 7
	} else {
		d.Counter--
	}
}

func (s *sweep) Clock() (r bool) {
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

func (e *envelope) Clock() {
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

func (t *timer) Clock() bool {
	if t.Tick == 0 {
		t.Tick = t.length
	} else {
		t.Tick--
	}
	return t.Tick == t.length
}

func (s *square) Clock() {
	if s.timer.Clock() {
		s.duty.Clock()
	}
}

func (t *triangle) Clock() {
	if t.timer.Clock() && t.length.Counter > 0 && t.linear.Counter > 0 {
		if t.SI == 31 {
			t.SI = 0
		} else {
			t.SI++
		}
	}
}

func (n *noise) Clock() {
	if n.timer.Clock() {
		var feedback uint16
		if n.Short {
			feedback = n.Shift & 0x40 << 8
		} else {
			feedback = n.Shift << 13
		}
		feedback ^= n.Shift << 14
		n.Shift >>= 1
		n.Shift &= 0x3fff
		n.Shift |= feedback
	}
}

func (a *apu) Step() {
	if a.Odd {
		if a.S1.Enable {
			a.S1.Clock()
		}
		if a.S2.Enable {
			a.S2.Clock()
		}
		if a.noise.Enable {
			a.noise.Clock()
		}
	}
	a.Odd = !a.Odd
	if a.triangle.Enable {
		a.triangle.Clock()
	}
}

func (a *apu) FrameStep() {
	a.FT++
	if a.FT == a.FC {
		a.FT = 0
	}
	if a.FT <= 3 {
		a.S1.envelope.Clock()
		a.S2.envelope.Clock()
		a.triangle.linear.Clock()
		a.noise.envelope.Clock()
	}
	if a.FT == 1 || a.FT == 3 {
		a.S1.FrameStep()
		a.S2.FrameStep()
		a.triangle.length.Clock()
		a.noise.length.Clock()
	}
	if a.FC == 4 && a.FT == 3 && !a.IrqDisable {
		a.Interrupt = true
	}
}

func (l *linear) Clock() {
	if l.Halt {
		l.Counter = l.Reload
	} else if l.Counter != 0 {
		l.Counter--
	}
	if !l.Flag {
		l.Halt = false
	}
}

func (s *square) FrameStep() {
	s.length.Clock()
	if s.sweep.Clock() && s.sweep.Enable && s.sweep.Shift > 0 {
		r := s.SweepResult()
		if r <= 0x7ff {
			s.timer.Tick = r
		}
	}
}

func (l *length) Clock() {
	if !l.Halt && l.Counter > 0 {
		l.Counter--
	}
}

func (a *apu) Volume() float32 {
	p := pulseOut[a.S1.Volume()+a.S2.Volume()]
	t := tndOut[3*a.triangle.Volume()+2*a.noise.Volume()]
	return p + t
}

func (n *noise) Volume() uint8 {
	if n.Enable && n.length.Counter > 0 && n.Shift&0x1 != 0 {
		return n.envelope.Output()
	}
	return 0
}

func (t *triangle) Volume() uint8 {
	if t.Enable && t.linear.Counter > 0 && t.length.Counter > 0 {
		return triLookup[t.SI]
	}
	return 0
}

func (s *square) Volume() uint8 {
	if s.Enable && s.duty.Enabled() && s.length.Enabled() && s.timer.Tick >= 8 && s.SweepResult() <= 0x7ff {
		return s.envelope.Output()
	}
	return 0
}

func (e *envelope) Output() byte {
	if e.Constant {
		return e.Volume
	}
	return e.Counter
}

func (s *square) SweepResult() uint16 {
	r := int(s.timer.Tick >> s.sweep.Shift)
	if s.sweep.Negate {
		r = -r
	}
	r += int(s.timer.Tick)
	if r > 0x7ff {
		r = 0x800
	}
	return uint16(r)
}

func (d *duty) Enabled() bool {
	return dutyCycle[d.Type][d.Counter] == 1
}

var (
	pulseOut  [31]float32
	tndOut    [203]float32
	dutyCycle = [4][8]byte{
		{0, 1, 0, 0, 0, 0, 0, 0},
		{0, 1, 1, 0, 0, 0, 0, 0},
		{0, 1, 1, 1, 1, 0, 0, 0},
		{1, 0, 0, 1, 1, 1, 1, 1},
	}
	lenLookup = [...]byte{
		0x0a, 0xfe, 0x14, 0x02,
		0x28, 0x04, 0x50, 0x06,
		0xa0, 0x08, 0x3c, 0x0a,
		0x0e, 0x0c, 0x1a, 0x0e,
		0x0c, 0x10, 0x18, 0x12,
		0x30, 0x14, 0x60, 0x16,
		0xc0, 0x18, 0x48, 0x1a,
		0x10, 0x1c, 0x20, 0x1e,
	}
	triLookup = [...]byte{
		0xF, 0xE, 0xD, 0xC,
		0xB, 0xA, 0x9, 0x8,
		0x7, 0x6, 0x5, 0x4,
		0x3, 0x2, 0x1, 0x0,
		0x0, 0x1, 0x2, 0x3,
		0x4, 0x5, 0x6, 0x7,
		0x8, 0x9, 0xA, 0xB,
		0xC, 0xD, 0xE, 0xF,
	}
	noiseLookup = [...]uint16{
		0x004, 0x008, 0x010, 0x020,
		0x040, 0x060, 0x080, 0x0a0,
		0x0ca, 0x0fe, 0x17c, 0x1fc,
		0x2fa, 0x3f8, 0x7f2, 0xfe4,
	}
)

func init() {
	for i := range pulseOut {
		pulseOut[i] = 95.88 / (8128/float32(i) + 100)
	}
	for i := range tndOut {
		tndOut[i] = 163.67 / (24329/float32(i) + 100)
	}
}
