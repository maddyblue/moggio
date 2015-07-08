package rardecode

import (
	"errors"
	"io"
)

const (
	rangeBottom = 1 << 15
	rangeTop    = 1 << 24

	maxFreq = 124

	intBits    = 7
	periodBits = 7
	binScale   = 1 << (intBits + periodBits)
)

var (
	errCorruptPPM = errors.New("rardecode: corrupt ppm data")

	expEscape  = []byte{25, 14, 9, 7, 5, 5, 4, 4, 4, 3, 3, 3, 2, 2, 2, 2}
	initBinEsc = []uint16{0x3CDD, 0x1F3F, 0x59BF, 0x48F3, 0x64A1, 0x5ABC, 0x6632, 0x6051}

	ns2Index   [256]byte
	ns2BSIndex [256]byte
)

func init() {
	ns2BSIndex[0] = 2 * 0
	ns2BSIndex[1] = 2 * 1
	for i := 2; i < 11; i++ {
		ns2BSIndex[i] = 2 * 2
	}
	for i := 11; i < 256; i++ {
		ns2BSIndex[i] = 2 * 3
	}

	var j, n byte
	for i := range ns2Index {
		ns2Index[i] = n
		if j <= 3 {
			n++
			j = n
		} else {
			j--
		}
	}
}

type rangeCoder struct {
	br   io.ByteReader
	code uint32
	low  uint32
	rnge uint32
}

func (r *rangeCoder) init(br io.ByteReader) error {
	r.br = br
	r.low = 0
	r.rnge = ^uint32(0)
	for i := 0; i < 4; i++ {
		c, err := r.br.ReadByte()
		if err != nil {
			return err
		}
		r.code = r.code<<8 | uint32(c)
	}
	return nil
}

func (r *rangeCoder) currentCount(scale uint32) uint32 {
	r.rnge /= scale
	return (r.code - r.low) / r.rnge
}

func (r *rangeCoder) normalize() error {
	for {
		if r.low^(r.low+r.rnge) >= rangeTop {
			if r.rnge >= rangeBottom {
				return nil
			}
			r.rnge = -r.low & (rangeBottom - 1)
		}
		c, err := r.br.ReadByte()
		if err != nil {
			return err
		}
		r.code = r.code<<8 | uint32(c)
		r.rnge <<= 8
		r.low <<= 8
	}
}

func (r *rangeCoder) decode(lowCount, highCount uint32) error {
	r.low += r.rnge * lowCount
	r.rnge *= highCount - lowCount

	return r.normalize()
}

type see2Context struct {
	summ  uint16
	shift byte
	count byte
}

func newSee2Context(i uint16) see2Context {
	return see2Context{i << (periodBits - 4), (periodBits - 4), 4}
}

func (s *see2Context) mean() uint32 {
	if s == nil {
		return 1
	}
	n := s.summ >> s.shift
	if n == 0 {
		return 1
	}
	s.summ -= n
	return uint32(n)
}

func (s *see2Context) update() {
	if s == nil || s.shift >= periodBits {
		return
	}
	s.count--
	if s.count == 0 {
		s.summ += s.summ
		s.count = 3 << s.shift
		s.shift++
	}
}

type state struct {
	sym  byte
	freq byte
	succ *context // successor
}

type context struct {
	summFreq uint16
	states   []state
	suffix   *context
}

type model struct {
	maxOrder    int
	orderFall   int
	initRL      int
	runLength   int
	prevSuccess byte
	escCount    byte
	prevSym     byte
	initEsc     byte
	minC        *context
	maxC        *context
	heapC       *context
	rc          rangeCoder
	charMask    [256]byte
	binSumm     [128][64]uint16
	see2Cont    [25][16]see2Context
}

func (m *model) restart() {
	for i := range m.charMask {
		m.charMask[i] = 0
	}
	m.escCount = 1

	if m.maxOrder < 12 {
		m.initRL = -m.maxOrder - 1
	} else {
		m.initRL = -12 - 1
	}
	m.orderFall = m.maxOrder
	m.runLength = m.initRL
	m.prevSuccess = 0

	c := new(context)
	c.summFreq = 257
	c.states = make([]state, 256)
	for i := range c.states {
		c.states[i].sym = byte(i)
		c.states[i].freq = 1
		c.states[i].succ = nil
	}
	m.minC = c
	m.maxC = c
	m.prevSym = 0
	m.heapC = new(context)

	for i := range m.binSumm {
		for j, esc := range initBinEsc {
			n := binScale - esc/(uint16(i)+2)
			for k := j; k < len(m.binSumm[i]); k += len(initBinEsc) {
				m.binSumm[i][k] = n
			}
		}
	}

	for i := range m.see2Cont {
		see := newSee2Context(5*uint16(i) + 10)
		for j := range m.see2Cont[i] {
			m.see2Cont[i][j] = see
		}
	}

}

func (m *model) init(br io.ByteReader, reset bool, maxOrder int) error {
	err := m.rc.init(br)
	if err != nil {
		return err
	}
	if !reset {
		if m.minC == nil {
			return errCorruptPPM
		}
		return nil
	}
	if maxOrder == 1 {
		return errCorruptPPM
	}
	m.maxOrder = maxOrder
	m.restart()
	return nil

}

func (m *model) rescale(s *state) *state {
	if s.freq <= maxFreq {
		return s
	}
	c := m.minC

	s.freq += 4
	c.summFreq += 4
	states := c.states
	escFreq := c.summFreq
	c.summFreq = 0
	for i := range states {
		f := states[i].freq
		escFreq -= uint16(f)
		if m.orderFall != 0 {
			f++
		}
		f >>= 1
		c.summFreq += uint16(f)
		states[i].freq = f

		if i == 0 || f <= states[i-1].freq {
			continue
		}
		j := i - 1
		for j > 0 && f > states[j-1].freq {
			j--
		}
		t := states[i]
		copy(states[j+1:i+1], states[j:i])
		states[j] = t
	}

	i := len(states) - 1
	for states[i].freq == 0 {
		i--
		escFreq++
	}
	c.states = states[:i+1]
	s = &states[0]
	if len(c.states) == 1 {
		for {
			s.freq -= s.freq >> 1
			escFreq >>= 1
			if escFreq <= 1 {
				return s
			}
		}
	}
	c.summFreq += escFreq - (escFreq >> 1)
	return s
}

func (m *model) decodeBinSymbol() (*state, error) {
	c := m.minC
	s := &c.states[0]

	i := m.prevSuccess + ns2BSIndex[len(c.suffix.states)-1] + byte(m.runLength>>26)&0x20
	if m.prevSym >= 64 {
		i += 8
	}
	if s.sym >= 64 {
		i += 2 * 8
	}
	bs := &m.binSumm[s.freq-1][i]
	mean := (*bs + 1<<(periodBits-2)) >> periodBits

	if m.rc.currentCount(binScale) < uint32(*bs) {
		err := m.rc.decode(0, uint32(*bs))
		if s.freq < 128 {
			s.freq++
		}
		*bs += 1<<intBits - mean
		m.prevSuccess = 1
		m.runLength++
		return s, err
	}
	err := m.rc.decode(uint32(*bs), binScale)
	*bs -= mean
	m.initEsc = expEscape[*bs>>10]
	m.charMask[s.sym] = m.escCount
	m.prevSuccess = 0
	return nil, err
}

func (m *model) decodeSymbol1() (*state, error) {
	c := m.minC
	states := c.states
	scale := uint32(c.summFreq)
	// protect against divide by zero
	// TODO: look at why this happens, may be problem elsewhere
	if scale == 0 {
		return nil, errCorruptPPM
	}
	count := m.rc.currentCount(scale)
	m.prevSuccess = 0

	var n uint32
	for i := range states {
		s := &states[i]
		n += uint32(s.freq)
		if n <= count {
			continue
		}
		err := m.rc.decode(n-uint32(s.freq), n)
		s.freq += 4
		c.summFreq += 4
		if i == 0 {
			if 2*n > scale {
				m.prevSuccess = 1
				m.runLength++
			}
		} else {
			if s.freq <= states[i-1].freq {
				return s, err
			}
			states[i-1], states[i] = states[i], states[i-1]
			s = &states[i-1]
		}
		return m.rescale(s), err
	}

	for _, s := range states {
		m.charMask[s.sym] = m.escCount
	}
	return nil, m.rc.decode(n, scale)
}

func (m *model) makeEscFreq(c *context, numMasked int) *see2Context {
	ns := len(c.states)
	if ns == 256 {
		return nil
	}
	diff := ns - numMasked

	var i int
	if m.prevSym >= 64 {
		i = 8
	}
	if diff < len(c.suffix.states)-ns {
		i++
	}
	if int(c.summFreq) < 11*ns {
		i += 2
	}
	if numMasked > diff {
		i += 4
	}
	return &m.see2Cont[ns2Index[diff-1]][i]
}

func (m *model) decodeSymbol2(numMasked int) (*state, error) {
	c := m.minC

	see := m.makeEscFreq(c, numMasked)
	scale := see.mean()

	var i int
	var hi uint32
	sl := make([]*state, len(c.states)-numMasked)
	for j := range sl {
		for m.charMask[c.states[i].sym] == m.escCount {
			i++
		}
		hi += uint32(c.states[i].freq)
		sl[j] = &c.states[i]
		i++
	}

	scale += hi
	count := m.rc.currentCount(scale)

	if count >= scale {
		return nil, errCorruptPPM
	}
	if count >= hi {
		err := m.rc.decode(hi, scale)
		if see != nil {
			see.summ += uint16(scale)
		}
		for _, s := range sl {
			m.charMask[s.sym] = m.escCount
		}
		return nil, err
	}

	hi = uint32(sl[0].freq)
	for hi <= count {
		sl = sl[1:]
		hi += uint32(sl[0].freq)
	}
	s := sl[0]

	err := m.rc.decode(hi-uint32(s.freq), hi)

	see.update()

	m.escCount++
	m.runLength = m.initRL

	s.freq += 4
	c.summFreq += 4
	return m.rescale(s), err
}

func (c *context) findState(sym byte) *state {
	var i int
	for i = range c.states {
		if c.states[i].sym == sym {
			break
		}
	}
	return &c.states[i]
}

func (m *model) createSuccessors(s, ss *state) *context {
	var sl []*state

	if m.orderFall != 0 {
		sl = append(sl, s)
	}

	c := m.minC
	for c.suffix != nil {
		c = c.suffix

		if ss == nil {
			ss = c.findState(s.sym)
		}
		if ss.succ != s.succ {
			c = ss.succ
			break
		}
		sl = append(sl, ss)
		ss = nil
	}

	if len(sl) == 0 {
		return c
	}

	var up state
	up.sym = byte(s.succ.summFreq) // get symbol from heap (context)
	up.succ = s.succ.suffix        // get next heap address (context)

	if len(c.states) > 1 {
		s = c.findState(up.sym)

		cf := uint16(s.freq) - 1
		s0 := c.summFreq - uint16(len(c.states)) - cf

		if 2*cf <= s0 {
			if 5*cf > s0 {
				up.freq = 2
			} else {
				up.freq = 1
			}
		} else {
			up.freq = byte(1 + (2*cf+3*s0-1)/(2*s0))
		}
	} else {
		up.freq = c.states[0].freq
	}

	for i := len(sl) - 1; i >= 0; i-- {
		c = &context{states: []state{up}, suffix: c}
		sl[i].succ = c
	}
	return c
}

func (m *model) update(s *state) {
	if m.escCount == 0 {
		m.escCount = 1
		for i := range m.charMask {
			m.charMask[i] = 0
		}
	}

	var ss *state // matching minC.suffix state

	if s.freq < maxFreq/4 && m.minC.suffix != nil {
		c := m.minC.suffix
		states := c.states

		var i int
		if len(states) > 1 {
			for states[i].sym != s.sym {
				i++
			}
			if i > 0 && states[i].freq >= states[i-1].freq {
				states[i-1], states[i] = states[i], states[i-1]
				i--
			}
			if states[i].freq < maxFreq-9 {
				states[i].freq += 2
				c.summFreq += 2
			}
		} else if states[0].freq < 32 {
			states[0].freq++
		}
		ss = &states[i] // save later for createSuccessors
	}

	if m.orderFall == 0 {
		c := m.createSuccessors(s, ss)
		m.minC = c
		m.maxC = c
		s.succ = c
		if c == nil {
			m.restart()
		}
		return
	}

	// Fake the heap by using a linked list of context's. Each context represents
	// an address, with the next address represented by the suffix context.
	// Data for that address is stored in the summFreq field.
	// The states slice is always nil for a heap context.

	succ := new(context)
	prevHeap := m.heapC
	prevHeap.summFreq = uint16(s.sym)
	prevHeap.suffix = succ
	m.heapC = succ

	minC := s.succ
	if minC == nil {
		s.succ = succ
		minC = m.minC
	} else {
		if minC.states == nil {
			minC = m.createSuccessors(s, ss)
			if minC == nil {
				m.restart()
				return
			}
		}
		m.orderFall--
		if m.orderFall == 0 {
			succ = minC
			if m.maxC != m.minC {
				prevHeap.suffix = nil
				m.heapC = prevHeap
			}
		}
	}

	n := len(m.minC.states)
	s0 := int(m.minC.summFreq) - n - int(s.freq-1)
	for c := m.maxC; c != m.minC; c = c.suffix {
		if ns := len(c.states); ns != 1 {
			if 4*ns <= n && int(c.summFreq) <= 8*ns {
				c.summFreq += 2
			}
			if 2*ns < n {
				c.summFreq++
			}
		} else {
			p := &c.states[0]
			if p.freq < maxFreq/4-1 {
				p.freq += p.freq
			} else {
				p.freq = maxFreq - 4
			}
			c.summFreq = uint16(p.freq) + uint16(m.initEsc)
			if n > 3 {
				c.summFreq++
			}
		}

		cf := 2 * int(s.freq) * int(c.summFreq+6)
		sf := s0 + int(c.summFreq)
		var freq byte
		if cf >= 6*sf {
			switch {
			case cf >= 15*sf:
				freq = 7
			case cf >= 12*sf:
				freq = 6
			case cf >= 9*sf:
				freq = 5
			default:
				freq = 4
			}
			c.summFreq += uint16(freq)
		} else {
			switch {
			case cf >= 4*sf:
				freq = 3
			case cf > sf:
				freq = 2
			default:
				freq = 1
			}
			c.summFreq += 3
		}
		c.states = append(c.states, state{s.sym, freq, succ})
	}
	m.minC = minC
	m.maxC = minC
}

func (m *model) ReadByte() (byte, error) {
	if m.minC == nil || m.minC.states == nil {
		return 0, errCorruptPPM
	}
	var s *state
	var err error
	if len(m.minC.states) == 1 {
		s, err = m.decodeBinSymbol()
	} else {
		s, err = m.decodeSymbol1()
	}
	for s == nil && err == nil {
		n := len(m.minC.states)
		for len(m.minC.states) == n {
			m.orderFall++
			m.minC = m.minC.suffix
			if m.minC == nil || m.minC.states == nil {
				return 0, errCorruptPPM
			}
		}
		s, err = m.decodeSymbol2(n)
	}
	if err != nil {
		return 0, err
	}

	if m.orderFall == 0 && s.succ != nil && s.succ.states != nil {
		m.minC = s.succ
		m.maxC = s.succ
	} else {
		m.update(s)
	}
	m.prevSym = s.sym
	return s.sym, nil
}
