/*
 * Copyright (c) 2014 Matt Jibson <matt.jibson@gmail.com>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

// Package cpu6502 implements a 6502 emulator.
package cpu6502

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

type timing map[Mode]int

type Instruction struct {
	F               Func
	Imm             byte
	ZP, ZPX, ZPY    byte
	ABS, ABSX, ABSY byte
	IND, INDX, INDY byte
	SNGL, BRA       byte
	TIM             timing
}

var Optable [0xff + 1]*Op

type Func func(*Cpu, byte, uint16, Mode)

type Op struct {
	Mode
	F Func
	T int
}

func (o *Op) String() string {
	n := runtime.FuncForPC(reflect.ValueOf(o.F).Pointer()).Name()
	n = n[strings.LastIndex(n, ".")+1:]
	return n
}

type Mode int

func (m Mode) Format() string {
	switch m {
	case MODE_IMM:
		return "#$%02[1]X"
	case MODE_ZP:
		return "$%02[2]X"
	case MODE_ZPX:
		return "$%02[3]X,X"
	case MODE_ZPY:
		return "$%02[3]X,Y"
	case MODE_ABS:
		return "$%04[2]X"
	case MODE_ABSX:
		return "$%04[3]X,X"
	case MODE_ABSY:
		return "$%04[3]X,Y"
	case MODE_IND:
		return "($%04[2]X)"
	case MODE_INDX:
		return "($%02[3]X,X)"
	case MODE_INDY:
		return "($%02[3]X),Y"
	case MODE_BRA:
		return "$%02[1]X"
	default:
		return ""
	}
}

const (
	MODE_IMM Mode = iota
	MODE_ZP
	MODE_ZPX
	MODE_ZPY
	MODE_ABS
	MODE_ABSX
	MODE_ABSY
	MODE_IND
	MODE_INDX
	MODE_INDY
	MODE_SNGL
	MODE_BRA

	IRQ   = 0xfffe
	RESET = 0xfffc
	NMI   = 0xfffa
)

type Memory interface {
	Read(uint16) byte
	Write(uint16, byte)
}

type Ticker interface {
	Tick()
}

type Cpu struct {
	Register
	M    Memory
	T    Ticker
	Halt bool

	DisableDecimal bool

	// If non nil, will record registers on each step.
	L     []Log
	LI    int // Log index
	Debug bool

	stepCycles int
}

func (c *Cpu) StringLog() string {
	s := ""
	o := c.LI
	for i := range c.L {
		li := (i + o) % len(c.L)
		s += fmt.Sprintf("\n%v", c.L[li])
	}
	return s
}

type Register struct {
	A, X, Y, S, P byte
	PC            uint16
}

type Log struct {
	R    Register
	O    *Op
	I    byte // instruction
	C    int  // cycles
	V, T uint16
	B    byte
}

func (l Log) String() string {
	m := l.O.Mode.Format()
	if m != "" {
		m = fmt.Sprintf(m, l.B, l.V, l.T)
	}
	return fmt.Sprintf("%04X: %02X %3v %-8s p=%08b s=%02X a=%02X x=%02X y=%02X v=%04X b=%02X t=%04X c=%d", l.R.PC, l.I, l.O, m, l.R.P, l.R.S, l.R.A, l.R.X, l.R.Y, l.V, l.B, l.T, l.C)
}

func New(m Memory) *Cpu {
	c := Cpu{
		Register: Register{
			S: 0xff,
			P: P_B | P_X | P_I,
		},
		M: m,
	}
	return &c
}

func (c *Cpu) Run() {
	for !c.Halt {
		c.Step()
	}
}

func (c *Cpu) Reset() {
	c.PC = uint16(c.M.Read(RESET+1))<<8 | uint16(c.M.Read(RESET))
}

func (c *Cpu) Tick(i int) {
	if i == 0 {
		panic("cpu6502: cannot tick for 0")
	}
	for ; i > 0; i-- {
		if c.T != nil {
			c.T.Tick()
		}
		c.stepCycles++
	}
}

func (c *Cpu) Step() {
	pc := c.PC
	c.stepCycles = 0
	inst := c.M.Read(c.PC)
	c.PC++
	o := Optable[inst]
	if o == nil {
		return
	}
	var b byte
	var v, t uint16
	switch o.Mode {
	case MODE_IMM, MODE_BRA:
		b = c.M.Read(c.PC)
		c.PC++
	case MODE_ZP:
		v = uint16(c.M.Read(c.PC))
		b = c.M.Read(v)
		c.PC++
	case MODE_ZPX:
		t = uint16(c.M.Read(c.PC))
		v = t + uint16(c.X)
		v &= 0xff
		b = c.M.Read(v)
		c.PC++
	case MODE_ZPY:
		t = uint16(c.M.Read(c.PC))
		v = t + uint16(c.Y)
		v &= 0xff
		b = c.M.Read(v)
		c.PC++
	case MODE_ABS:
		v = uint16(c.M.Read(c.PC))
		c.PC++
		v |= uint16(c.M.Read(c.PC)) << 8
		c.PC++
		b = c.M.Read(v)
	case MODE_ABSX:
		t = uint16(c.M.Read(c.PC))
		c.PC++
		t |= uint16(c.M.Read(c.PC)) << 8
		c.PC++
		v = t + uint16(c.X)
		b = c.M.Read(v)
	case MODE_ABSY:
		t = uint16(c.M.Read(c.PC))
		c.PC++
		t |= uint16(c.M.Read(c.PC)) << 8
		c.PC++
		v = t + uint16(c.Y)
		b = c.M.Read(v)
	case MODE_IND:
		v = uint16(c.M.Read(c.PC))
		c.PC++
		v |= uint16(c.M.Read(c.PC)) << 8
		v = uint16(c.M.Read(v)) + uint16(c.M.Read(v+1))<<8
		c.PC++
	case MODE_INDX:
		t = uint16(c.M.Read(c.PC))
		c.PC++
		v = t + uint16(c.X)
		v &= 0xff
		v1 := v + 1
		v1 &= 0xff
		v = uint16(c.M.Read(v)) + uint16(c.M.Read(v1))<<8
		b = c.M.Read(v)
	case MODE_INDY:
		t = uint16(c.M.Read(c.PC))
		c.PC++
		t1 := t + 1
		t1 &= 0xff
		v = uint16(c.M.Read(t)) + uint16(c.M.Read(t1))<<8 + uint16(c.Y)
		b = c.M.Read(v)
	case MODE_SNGL:
		// nothing
	default:
		panic("6502: bad address mode")
	}
	o.F(c, b, v, o.Mode)
	c.Tick(o.T)
	if c.L != nil || c.Debug {
		r := c.Register
		r.PC = pc
		l := Log{
			R: r,
			O: o,
			I: inst,
			C: c.stepCycles,
			V: v,
			T: t,
			B: b,
		}
		if c.L != nil {
			c.L[c.LI] = l
			c.LI++
			c.LI %= len(c.L)
		}
		if c.Debug {
			fmt.Println(l)
		}
	}
}

func (c *Cpu) setNV(v byte) {
	if v != 0 {
		c.P &= ^P_Z
	} else {
		c.P |= P_Z
	}
	if v&0x80 != 0 {
		c.P |= P_N
	} else {
		c.P &= ^P_N
	}
}

func (c *Cpu) SEC() { c.P |= P_C }
func (c *Cpu) CLC() { c.P &= ^P_C }
func (c *Cpu) SEV() { c.P |= P_V }
func (c *Cpu) CLV() { c.P &= ^P_V }
func (c *Cpu) SEI() { c.P |= P_I }
func (c *Cpu) CLI() { c.P &= ^P_I }
func (c *Cpu) SED() { c.P |= P_D }
func (c *Cpu) CLD() { c.P &= ^P_D }

func (c *Cpu) C() bool       { return c.p(P_C) }
func (c *Cpu) Z() bool       { return c.p(P_Z) }
func (c *Cpu) I() bool       { return c.p(P_I) }
func (c *Cpu) D() bool       { return c.p(P_D) }
func (c *Cpu) B() bool       { return c.p(P_B) }
func (c *Cpu) V() bool       { return c.p(P_V) }
func (c *Cpu) N() bool       { return c.p(P_N) }
func (c *Cpu) p(v byte) bool { return c.P&v != 0 }

const (
	P_C byte = 1 << iota
	P_Z
	P_I
	P_D
	P_B
	P_X // unused
	P_V
	P_N
)

func (c *Cpu) String() string {
	const f = "%2s: %5d 0x%04[2]X %016[2]b\n"
	s := "\n"
	s += fmt.Sprintf(f, "A", c.A)
	s += fmt.Sprintf(f, "X", c.X)
	s += fmt.Sprintf(f, "Y", c.Y)
	s += fmt.Sprintf(f, "P", c.P)
	s += fmt.Sprintf(f, "S", c.S)
	s += fmt.Sprintf(f, "PC", c.PC)
	return s
}

func init() {
	populate := func(i Instruction, m Mode, v byte) {
		if v != null {
			if Optable[v] != nil {
				panic("duplicate instruction")
			} else if i.TIM[m] == 0 {
				panic("no timing information")
			}
			Optable[v] = &Op{
				F:    i.F,
				Mode: m,
				T:    i.TIM[m],
			}
		}
	}
	for _, i := range Opcodes {
		populate(i, MODE_IMM, i.Imm)
		populate(i, MODE_ZP, i.ZP)
		populate(i, MODE_ZPX, i.ZPX)
		populate(i, MODE_ZPY, i.ZPY)
		populate(i, MODE_ABS, i.ABS)
		populate(i, MODE_ABSX, i.ABSX)
		populate(i, MODE_ABSY, i.ABSY)
		populate(i, MODE_IND, i.IND)
		populate(i, MODE_INDX, i.INDX)
		populate(i, MODE_INDY, i.INDY)
		populate(i, MODE_SNGL, i.SNGL)
		populate(i, MODE_BRA, i.BRA)
	}
	Optable[0] = &Op{
		F:    BRK,
		Mode: MODE_BRA,
		T:    _K[MODE_BRA],
	}
}

func (c *Cpu) Interrupt() {
	a := uint16(c.M.Read(NMI)) + uint16(c.M.Read(NMI+1))<<8
	if a == 0 {
		panic("BAD NMI")
		c.Halt = true
		return
	}
	c.stackPush(byte(c.PC >> 8))
	c.stackPush(byte(c.PC & 0xff))
	c.stackPush(c.P | P_X | P_B)
	c.P |= P_I
	c.Tick(Optable[0].T)
}

func BRK(c *Cpu, b byte, v uint16, m Mode) {
	a := uint16(c.M.Read(IRQ)) + uint16(c.M.Read(IRQ+1))<<8
	if a == 0 {
		c.Halt = true
		return
	}
	c.stackPush(byte(c.PC >> 8))
	c.stackPush(byte(c.PC & 0xff))
	c.stackPush(c.P | P_B)
	c.PC = a
	c.P |= P_I
}

func NOP(c *Cpu, b byte, v uint16, m Mode) {}

func ADC(c *Cpu, b byte, v uint16, m Mode) {
	if (c.A^b)&0x80 != 0 {
		c.CLV()
	} else {
		c.SEV()
	}
	var a uint16
	if c.D() && !c.DisableDecimal {
		a = uint16(c.A&0xf) + uint16(b&0xf)
		if c.C() {
			a++
		}
		if a >= 10 {
			a = 0x10 | (a+6)&0xf
		}
		a += uint16(c.A&0xf0) + uint16(b&0xf0)
		if a >= 160 {
			c.SEC()
			if c.V() && a >= 0x180 {
				c.CLV()
			}
			a += 0x60
		} else {
			c.CLC()
			if c.V() && a < 0x80 {
				c.CLV()
			}
		}
	} else {
		a = uint16(c.A) + uint16(b)
		if c.C() {
			a++
		}
		if a > 0xff {
			c.SEC()
			if c.V() && a >= 0x180 {
				c.CLV()
			}
		} else {
			c.CLC()
			if c.V() && a < 0x80 {
				c.CLV()
			}
		}
	}
	c.A = byte(a & 0xff)
	c.setNV(c.A)
}

func SBC(c *Cpu, b byte, v uint16, m Mode) {
	if (c.A^b)&0x80 != 0 {
		c.SEV()
	} else {
		c.CLV()
	}
	var a uint16
	if c.D() && !c.DisableDecimal {
		var w uint16
		a = 0xf + uint16(c.A&0xf) - uint16(b&0xf)
		if c.C() {
			a++
		}
		if a < 0x10 {
			a -= 6
		} else {
			w = 0x10
			a -= 0x10
		}
		w += 0xf0 + uint16(c.A&0xf0) - uint16(b&0xf0)
		if w < 0x100 {
			c.CLC()
			if c.V() && w < 0x80 {
				c.CLV()
			}
			w -= 0x60
		} else {
			c.SEC()
			if c.V() && w >= 0x180 {
				c.CLV()
			}
		}
		a += w
	} else {
		a = 0xff + uint16(c.A) - uint16(b)
		if c.C() {
			a++
		}
		if a < 0x100 {
			c.CLC()
			if c.V() && a < 0x80 {
				c.CLV()
			}
		} else {
			c.SEC()
			if c.V() && a >= 0x180 {
				c.CLV()
			}
		}
	}
	c.A = byte(a & 0xff)
	c.setNV(c.A)
}

func LDA(c *Cpu, b byte, v uint16, m Mode) {
	c.A = b
	c.setNV(c.A)
}

func LDX(c *Cpu, b byte, v uint16, m Mode) {
	c.X = b
	c.setNV(c.X)
}

func LDY(c *Cpu, b byte, v uint16, m Mode) {
	c.Y = b
	c.setNV(c.Y)
}

func STA(c *Cpu, b byte, v uint16, m Mode) {
	c.M.Write(v, c.A)
}

func STX(c *Cpu, b byte, v uint16, m Mode) {
	c.M.Write(v, c.X)
}

func STY(c *Cpu, b byte, v uint16, m Mode) {
	c.M.Write(v, c.Y)
}

func TAX(c *Cpu, b byte, v uint16, m Mode) {
	c.X = c.A
	c.setNV(c.X)
}

func TAY(c *Cpu, b byte, v uint16, m Mode) {
	c.Y = c.A
	c.setNV(c.Y)
}

func TYA(c *Cpu, b byte, v uint16, m Mode) {
	c.A = c.Y
	c.setNV(c.A)
}

func TXA(c *Cpu, b byte, v uint16, m Mode) {
	c.A = c.X
	c.setNV(c.A)
}

func TSX(c *Cpu, b byte, v uint16, m Mode) {
	c.X = c.S
	c.setNV(c.X)
}

func TXS(c *Cpu, b byte, v uint16, m Mode) {
	c.S = c.X
}

func INX(c *Cpu, b byte, v uint16, m Mode) {
	c.X = (c.X + 1) & 0xff
	c.setNV(c.X)
}

func INY(c *Cpu, b byte, v uint16, m Mode) {
	c.Y = (c.Y + 1) & 0xff
	c.setNV(c.Y)
}

func INC(c *Cpu, b byte, v uint16, m Mode) {
	c.M.Write(v, (c.M.Read(v)+1)&0xff)
	c.setNV(c.M.Read(v))
}

func DEX(c *Cpu, b byte, v uint16, m Mode) {
	c.X = (c.X - 1) & 0xff
	c.setNV(c.X)
}

func DEY(c *Cpu, b byte, v uint16, m Mode) {
	c.Y = (c.Y - 1) & 0xff
	c.setNV(c.Y)
}

func DEC(c *Cpu, b byte, v uint16, m Mode) {
	c.M.Write(v, (c.M.Read(v)-1)&0xff)
	c.setNV(c.M.Read(v))
}

func CMP(c *Cpu, b byte, v uint16, m Mode) { c.compare(c.A, b) }
func CPX(c *Cpu, b byte, v uint16, m Mode) { c.compare(c.X, b) }
func CPY(c *Cpu, b byte, v uint16, m Mode) { c.compare(c.Y, b) }

func (c *Cpu) compare(r, v byte) {
	if r >= v {
		c.SEC()
	} else {
		c.CLC()
	}
	c.setNV(r - v)
}

func BCC(c *Cpu, b byte, v uint16, m Mode) {
	if !c.C() {
		c.jump(uint16(b))
	}
}

func BCS(c *Cpu, b byte, v uint16, m Mode) {
	if c.C() {
		c.jump(uint16(b))
	}
}

func BNE(c *Cpu, b byte, v uint16, m Mode) {
	if !c.Z() {
		c.jump(uint16(b))
	}
}

func BEQ(c *Cpu, b byte, v uint16, m Mode) {
	if c.Z() {
		c.jump(uint16(b))
	}
}

func BPL(c *Cpu, b byte, v uint16, m Mode) {
	if !c.N() {
		c.jump(uint16(b))
	}
}

func BMI(c *Cpu, b byte, v uint16, m Mode) {
	if c.N() {
		c.jump(uint16(b))
	}
}

func BVC(c *Cpu, b byte, v uint16, m Mode) {
	if !c.V() {
		c.jump(uint16(b))
	}
}

func BVS(c *Cpu, b byte, v uint16, m Mode) {
	if c.V() {
		c.jump(uint16(b))
	}
}

func (c *Cpu) jump(v uint16) {
	c.Tick(1)
	if v > 0x7f {
		c.PC -= 0x100 - v
	} else {
		c.PC += v
	}
}

func JMP(c *Cpu, b byte, v uint16, m Mode) {
	c.PC = uint16(v)
}

func PHA(c *Cpu, b byte, v uint16, m Mode) {
	c.stackPush(c.A)
}

func PLA(c *Cpu, b byte, v uint16, m Mode) {
	c.A = c.stackPop()
	c.setNV(c.A)
}

func (c *Cpu) stackPush(b byte) {
	c.M.Write(uint16(c.S)+0x100, b)
	c.S = (c.S - 1) & 0xff
}

func (c *Cpu) stackPop() byte {
	c.S = (c.S + 1) & 0xff
	return c.M.Read(uint16(c.S) + 0x100)
}

func JSR(c *Cpu, b byte, v uint16, m Mode) {
	a := c.PC - 1
	c.stackPush(byte(a >> 8))
	c.stackPush(byte(a & 0xff))
	c.PC = v
}

func RTS(c *Cpu, b byte, v uint16, m Mode) {
	c.PC = (uint16(c.stackPop()) | uint16(c.stackPop())<<8)
	if c.PC == 0 {
		c.Halt = true
	} else {
		c.PC++
	}
}

func AND(c *Cpu, b byte, v uint16, m Mode) {
	c.A &= b
	c.setNV(c.A)
}

func ORA(c *Cpu, b byte, v uint16, m Mode) {
	c.A |= b
	c.setNV(c.A)
}

func ASL(c *Cpu, b byte, v uint16, m Mode) {
	if m == MODE_SNGL {
		c.setCarryBit(c.A, 7)
		c.A <<= 1
		c.setNV(c.A)
	} else {
		c.setCarryBit(c.M.Read(v), 7)
		c.M.Write(v, c.M.Read(v)<<1)
		c.setNV(c.M.Read(v))
	}
}

func ROL(c *Cpu, b byte, v uint16, m Mode) {
	var s byte
	if c.C() {
		s = 0x01
	}
	if m == MODE_SNGL {
		c.setCarryBit(c.A, 7)
		c.A <<= 1
		c.A |= s
		c.setNV(c.A)
	} else {
		c.setCarryBit(c.M.Read(v), 7)
		c.M.Write(v, c.M.Read(v)<<1)
		c.M.Write(v, c.M.Read(v)|s)
		c.setNV(c.M.Read(v))
	}
}

func LSR(c *Cpu, b byte, v uint16, m Mode) {
	if m == MODE_SNGL {
		c.setCarryBit(c.A, 0)
		c.A >>= 1
		c.setNV(c.A)
	} else {
		c.setCarryBit(c.M.Read(v), 0)
		c.M.Write(v, c.M.Read(v)>>1)
		c.setNV(c.M.Read(v))
	}
}

func ROR(c *Cpu, b byte, v uint16, m Mode) {
	var s byte
	if c.C() {
		s = 0x80
	}
	if m == MODE_SNGL {
		c.setCarryBit(c.A, 0)
		c.A >>= 1
		c.A |= s
		c.setNV(c.A)
	} else {
		c.setCarryBit(c.M.Read(v), 0)
		c.M.Write(v, c.M.Read(v)>>1)
		c.M.Write(v, c.M.Read(v)|s)
		c.setNV(c.M.Read(v))
	}
}

func BIT(c *Cpu, b byte, v uint16, m Mode) {
	if b&0x80 != 0 {
		c.P |= P_N
	} else {
		c.P &= ^P_N
	}
	if b&0x40 != 0 {
		c.P |= P_V
	} else {
		c.P &= ^P_V
	}
	if c.A&b != 0 {
		c.P &= ^P_Z
	} else {
		c.P |= P_Z
	}
}

func CLC(c *Cpu, b byte, v uint16, m Mode) {
	c.CLC()
}

func SEC(c *Cpu, b byte, v uint16, m Mode) {
	c.SEC()
}

func CLI(c *Cpu, b byte, v uint16, m Mode) {
	c.CLI()
}

func SEI(c *Cpu, b byte, v uint16, m Mode) {
	c.SEI()
}

func CLD(c *Cpu, b byte, v uint16, m Mode) {
	c.CLD()
}

func SED(c *Cpu, b byte, v uint16, m Mode) {
	c.SED()
}

func CLV(c *Cpu, b byte, v uint16, m Mode) {
	c.CLV()
}

func (c *Cpu) setCarryBit(b byte, i uint) {
	if b>>i&0x01 != 0 {
		c.P |= P_C
	} else {
		c.P &= ^P_C
	}
}

func EOR(c *Cpu, b byte, v uint16, m Mode) {
	c.A ^= b
	c.setNV(c.A)
}

func PHP(c *Cpu, b byte, v uint16, m Mode) {
	c.stackPush(c.P | P_X | P_B)
}

func PLP(c *Cpu, b byte, v uint16, m Mode) {
	c.P = c.stackPop() | P_X
	c.P &= ^P_B
}

func RTI(c *Cpu, b byte, v uint16, m Mode) {
	c.P = c.stackPop() | P_X
	c.PC = uint16(c.stackPop()) + uint16(c.stackPop())<<8
}

func TRB(c *Cpu, b byte, v uint16, m Mode) {
	if c.A&c.M.Read(v) != 0 {
		c.P &= ^P_Z
	} else {
		c.P |= P_Z
	}
	c.M.Write(v, c.M.Read(v) & ^c.A)
}

func TSB(c *Cpu, b byte, v uint16, m Mode) {
	if c.A&c.M.Read(v) != 0 {
		c.P &= ^P_Z
	} else {
		c.P |= P_Z
	}
	c.M.Write(v, c.M.Read(v)|c.A)
}

const null = 0

var (
	_1 = timing{
		MODE_IMM:  2,
		MODE_ZP:   3,
		MODE_ZPX:  4,
		MODE_ZPY:  4,
		MODE_ABS:  4,
		MODE_ABSX: 4,
		MODE_ABSY: 4,
		MODE_INDX: 6,
		MODE_INDY: 5,
	}
	_2 = timing{
		MODE_SNGL: 2,
		MODE_IMM:  2,
		MODE_ZP:   5,
		MODE_ZPX:  6,
		MODE_ABS:  6,
		MODE_ABSX: 7,
	}
	_3 = timing{
		MODE_IMM:  2,
		MODE_ZP:   3,
		MODE_ZPX:  4,
		MODE_ZPY:  4,
		MODE_ABS:  4,
		MODE_ABSX: 5,
		MODE_ABSY: 5,
		MODE_INDX: 6,
		MODE_INDY: 6,
	}
	_B = timing{
		MODE_BRA: 2,
	}
	_S = timing{
		MODE_SNGL: 2,
	}
	_S3 = timing{
		MODE_SNGL: 3,
	}
	_S4 = timing{
		MODE_SNGL: 4,
	}
	_S6 = timing{
		MODE_SNGL: 6,
	}
	_K = timing{
		MODE_BRA: 7,
	}
	_J = timing{
		MODE_ABS: 3,
		MODE_IND: 5,
	}
)

var Opcodes = []Instruction{
	/* F,  Imm,   ZP,  ZPX,  ZPY,  ABS, ABSX, ABSY,  IND, INDX, INDY, SNGL,  BRA, TIM */
	{ADC, 0x69, 0x65, 0x75, null, 0x6d, 0x7d, 0x79, null, 0x61, 0x71, null, null, _1},
	{AND, 0x29, 0x25, 0x35, null, 0x2d, 0x3d, 0x39, null, 0x21, 0x31, null, null, _1},
	{ASL, null, 0x06, 0x16, null, 0x0e, 0x1e, null, null, null, null, 0x0a, null, _2},
	{BCC, null, null, null, null, null, null, null, null, null, null, null, 0x90, _B},
	{BCS, null, null, null, null, null, null, null, null, null, null, null, 0xb0, _B},
	{BEQ, null, null, null, null, null, null, null, null, null, null, null, 0xf0, _B},
	{BIT, null, 0x24, null, null, 0x2c, null, null, null, null, null, null, null, _3},
	{BMI, null, null, null, null, null, null, null, null, null, null, null, 0x30, _B},
	{BNE, null, null, null, null, null, null, null, null, null, null, null, 0xd0, _B},
	{BPL, null, null, null, null, null, null, null, null, null, null, null, 0x10, _B},
	{BRK, null, null, null, null, null, null, null, null, null, null, null, 0x00, _K},
	{BVC, null, null, null, null, null, null, null, null, null, null, null, 0x50, _B},
	{BVS, null, null, null, null, null, null, null, null, null, null, null, 0x70, _B},
	{CLC, null, null, null, null, null, null, null, null, null, null, 0x18, null, _S},
	{CLD, null, null, null, null, null, null, null, null, null, null, 0xd8, null, _S},
	{CLI, null, null, null, null, null, null, null, null, null, null, 0x58, null, _S},
	{CLV, null, null, null, null, null, null, null, null, null, null, 0xb8, null, _S},
	{CMP, 0xc9, 0xc5, 0xd5, null, 0xcd, 0xdd, 0xd9, null, 0xc1, 0xd1, null, null, _1},
	{CPX, 0xe0, 0xe4, null, null, 0xec, null, null, null, null, null, null, null, _2},
	{CPY, 0xc0, 0xc4, null, null, 0xcc, null, null, null, null, null, null, null, _2},
	{DEC, null, 0xc6, 0xd6, null, 0xce, 0xde, null, null, null, null, null, null, _2},
	{DEX, null, null, null, null, null, null, null, null, null, null, 0xca, null, _S},
	{DEY, null, null, null, null, null, null, null, null, null, null, 0x88, null, _S},
	{EOR, 0x49, 0x45, 0x55, null, 0x4d, 0x5d, 0x59, null, 0x41, 0x51, null, null, _1},
	{INC, null, 0xe6, 0xf6, null, 0xee, 0xfe, null, null, null, null, null, null, _2},
	{INX, null, null, null, null, null, null, null, null, null, null, 0xe8, null, _S},
	{INY, null, null, null, null, null, null, null, null, null, null, 0xc8, null, _S},
	{JMP, null, null, null, null, 0x4c, null, null, 0x6c, null, null, null, null, _J},
	{JSR, null, null, null, null, 0x20, null, null, null, null, null, null, null, _2},
	{LDA, 0xa9, 0xa5, 0xb5, null, 0xad, 0xbd, 0xb9, null, 0xa1, 0xb1, null, null, _1},
	{LDX, 0xa2, 0xa6, null, 0xb6, 0xae, null, 0xbe, null, null, null, null, null, _1},
	{LDY, 0xa0, 0xa4, 0xb4, null, 0xac, 0xbc, null, null, null, null, null, null, _1},
	{LSR, null, 0x46, 0x56, null, 0x4e, 0x5e, null, null, null, null, 0x4a, null, _2},
	{NOP, null, null, null, null, null, null, null, null, null, null, 0xea, null, _S},
	{ORA, 0x09, 0x05, 0x15, null, 0x0d, 0x1d, 0x19, null, 0x01, 0x11, null, null, _1},
	{PHA, null, null, null, null, null, null, null, null, null, null, 0x48, null, _S3},
	{PHP, null, null, null, null, null, null, null, null, null, null, 0x08, null, _S3},
	{PLA, null, null, null, null, null, null, null, null, null, null, 0x68, null, _S4},
	{PLP, null, null, null, null, null, null, null, null, null, null, 0x28, null, _S4},
	{ROL, null, 0x26, 0x36, null, 0x2e, 0x3e, null, null, null, null, 0x2a, null, _2},
	{ROR, null, 0x66, 0x76, null, 0x6e, 0x7e, null, null, null, null, 0x6a, null, _2},
	{RTI, null, null, null, null, null, null, null, null, null, null, 0x40, null, _S6},
	{RTS, null, null, null, null, null, null, null, null, null, null, 0x60, null, _S6},
	{SBC, 0xe9, 0xe5, 0xf5, null, 0xed, 0xfd, 0xf9, null, 0xe1, 0xf1, null, null, _1},
	{SEC, null, null, null, null, null, null, null, null, null, null, 0x38, null, _S},
	{SED, null, null, null, null, null, null, null, null, null, null, 0xf8, null, _S},
	{SEI, null, null, null, null, null, null, null, null, null, null, 0x78, null, _S},
	{STA, null, 0x85, 0x95, null, 0x8d, 0x9d, 0x99, null, 0x81, 0x91, null, null, _3},
	{STX, null, 0x86, null, 0x96, 0x8e, null, null, null, null, null, null, null, _3},
	{STY, null, 0x84, 0x94, null, 0x8c, null, null, null, null, null, null, null, _3},
	{TAX, null, null, null, null, null, null, null, null, null, null, 0xaa, null, _S},
	{TAY, null, null, null, null, null, null, null, null, null, null, 0xa8, null, _S},
	{TRB, null, 0x14, null, null, 0x1c, null, null, null, null, null, null, null, _2},
	{TSB, null, 0x04, null, null, 0x0c, null, null, null, null, null, null, null, _2},
	{TSX, null, null, null, null, null, null, null, null, null, null, 0xba, null, _S},
	{TXA, null, null, null, null, null, null, null, null, null, null, 0x8a, null, _S},
	{TXS, null, null, null, null, null, null, null, null, null, null, 0x9a, null, _S},
	{TYA, null, null, null, null, null, null, null, null, null, null, 0x98, null, _S},
}
