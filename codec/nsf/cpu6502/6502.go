package cpu6502

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

type Instruction struct {
	F               Func
	Imm             byte
	ZP, ZPX, ZPY    byte
	ABS, ABSX, ABSY byte
	IND, INDX, INDY byte
	SNGL, BRA       byte
}

var Optable [0xff]*Op

type Func func(*Cpu, byte, uint16, Mode)

type Op struct {
	Mode
	F Func
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
		return "#$%02[1]x"
	case MODE_ZP:
		return "$%02[2]x"
	case MODE_ZPX:
		return "$%02[2]x,X"
	case MODE_ZPY:
		return "$%02[2]x,Y"
	case MODE_ABS:
		return "$%04[2]x"
	case MODE_ABSX:
		return "$%04[2]x,X"
	case MODE_ABSY:
		return "$%04[2]x,Y"
	case MODE_IND:
		return "($%04[2]X)"
	case MODE_INDX:
		return "($%02[2]X,X)"
	case MODE_INDY:
		return "($%02[2]X),Y"
	case MODE_BRA:
		return "$%02[1]x"
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
)

type Cpu struct {
	A, X, Y, S, P byte
	PC            uint16
	Mem           [0xffff]byte
	Halt          bool
}

func New() *Cpu {
	c := Cpu{
		S:  0xff,
		P:  0x30,
		PC: 0x0600,
	}
	return &c
}

func (c *Cpu) Run() {
	for !c.Halt {
		c.Step()
	}
}

func (c *Cpu) Step() {
	pc := c.PC
	inst := c.Mem[c.PC]
	c.PC++
	if inst == 0 {
		c.Halt = true
		return
	}
	o := Optable[inst]
	if o == nil {
		panic(fmt.Sprintf("bad opcode 0x%02x", inst))
		return
	}
	var b byte
	var v uint16
	switch o.Mode {
	case MODE_IMM, MODE_BRA:
		b = c.Mem[c.PC]
		c.PC++
	case MODE_ZP:
		v = uint16(c.Mem[c.PC])
		b = c.Mem[v]
		c.PC++
	case MODE_ZPX:
		v = uint16(c.Mem[c.PC])
		t := v + uint16(c.X)
		t &= 0xff
		b = c.Mem[t]
		c.PC++
	case MODE_ZPY:
		v = uint16(c.Mem[c.PC])
		t := v + uint16(c.Y)
		t &= 0xff
		b = c.Mem[t]
		c.PC++
	case MODE_ABS:
		v = uint16(c.Mem[c.PC])
		c.PC++
		v |= uint16(c.Mem[c.PC]) << 8
		c.PC++
		b = c.Mem[v]
	case MODE_ABSX:
		v = uint16(c.Mem[c.PC])
		c.PC++
		v |= uint16(c.Mem[c.PC]) << 8
		c.PC++
		v += uint16(c.X)
		b = c.Mem[v]
	case MODE_ABSY:
		v = uint16(c.Mem[c.PC])
		c.PC++
		v |= uint16(c.Mem[c.PC]) << 8
		c.PC++
		v += uint16(c.Y)
		b = c.Mem[v]
	case MODE_IND:
		v = uint16(c.Mem[c.PC])
		c.PC++
		v |= uint16(c.Mem[c.PC]) << 8
		v = uint16(c.Mem[v]) + uint16(c.Mem[v+1])<<8
		c.PC++
	case MODE_INDX:
		v = uint16(c.Mem[c.PC])
		c.PC++
		t := v + uint16(c.X)
		t &= 0xff
		t = uint16(c.Mem[t]) + uint16(c.Mem[t+1])<<8
		b = c.Mem[t]
	case MODE_INDY:
		v = uint16(c.Mem[c.PC])
		c.PC++
		t := uint16(c.Mem[v]) + uint16(c.Mem[v+1])<<8 + uint16(c.Y)
		b = c.Mem[t]
	case MODE_SNGL:
		// nothing
	default:
		panic("6502: bad address mode")
	}
	m := o.Mode.Format()
	if m != "" {
		m = fmt.Sprintf(m, b, v)
	}
	fmt.Printf("PC: 0x%04X, inst: 0x%02X %v %s\n", pc, inst, o, m)
	_ = pc
	o.F(c, b, v, o.Mode)
}

func (c *Cpu) setNV(v byte) {
	if v != 0 {
		c.P &= 0xfd
	} else {
		c.P |= 0x02
	}
	if v&0x80 != 0 {
		c.P |= 0x80
	} else {
		c.P &= 0x7f
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
func (c *Cpu) V() bool       { return c.p(P_V) }
func (c *Cpu) N() bool       { return c.p(P_N) }
func (c *Cpu) p(v byte) bool { return c.P&v != 0 }

const (
	P_C byte = 1 << iota
	P_Z
	P_I
	P_D
	P_B
	_
	P_V
	P_N
)

func (c *Cpu) String() string {
	const f = "%2s: %5d 0x%04[2]X %016[2]b\n"
	s := ""
	s += fmt.Sprintf(f, "A", c.A)
	s += fmt.Sprintf(f, "X", c.X)
	s += fmt.Sprintf(f, "Y", c.Y)
	s += fmt.Sprintf(f, "P", c.P)
	s += fmt.Sprintf(f, "PC", c.PC)
	return s
}

func init() {
	populate := func(i Instruction, m Mode, v byte) {
		if v != null {
			if Optable[v] != nil {
				panic("duplicate instruction")
			}
			Optable[v] = &Op{
				F:    i.F,
				Mode: m,
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
}

func BRK(c *Cpu, b byte, v uint16, m Mode) {}
func NOP(c *Cpu, b byte, v uint16, m Mode) {}

func ADC(c *Cpu, b byte, v uint16, m Mode) {
	if (c.A^b)&0x80 != 0 {
		c.CLV()
	} else {
		c.SEV()
	}
	a := uint16(c.A) + uint16(b)
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
	c.A = byte(a & 0xff)
	c.setNV(c.A)
}

func SBC(c *Cpu, b byte, v uint16, m Mode) {
	if (c.A^b)&0x80 != 0 {
		c.SEV()
	} else {
		c.CLV()
	}
	a := 0xff + uint16(c.A) - uint16(b)
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

func STA(c *Cpu, b byte, v uint16, m Mode) { c.Mem[v] = c.A }
func STX(c *Cpu, b byte, v uint16, m Mode) { c.Mem[v] = c.X }
func STY(c *Cpu, b byte, v uint16, m Mode) { c.Mem[v] = c.Y }

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
	c.Mem[v] = (c.Mem[v] + 1) & 0xff
	c.setNV(c.Mem[v])
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
	c.Mem[v] = (c.Mem[v] - 1) & 0xff
	c.setNV(c.Mem[v])
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
	c.Mem[uint16(c.S)+0x100] = b
	if c.S == 0 {
		// wrap
		c.S = 0xff
	} else {
		c.S--
	}
}

func (c *Cpu) stackPop() byte {
	if c.S < 0xff {
		c.S++
	}
	return c.Mem[uint16(c.S)+0x100]
}

func JSR(c *Cpu, b byte, v uint16, m Mode) {
	a := c.PC - 1
	c.stackPush(byte(a >> 8))
	c.stackPush(byte(a & 0xff))
	c.PC = v
}

func RTS(c *Cpu, b byte, v uint16, m Mode) {
	c.PC = (uint16(c.stackPop()) | uint16(c.stackPop())<<8) + 1
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
		c.setCarryBit(c.Mem[v], 7)
		c.Mem[v] <<= 1
		c.setNV(c.Mem[v])
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
		c.setCarryBit(c.Mem[v], 7)
		c.Mem[v] <<= 1
		c.Mem[v] |= s
		c.setNV(c.Mem[v])
	}
}

func LSR(c *Cpu, b byte, v uint16, m Mode) {
	if m == MODE_SNGL {
		c.A >>= 1
		c.setNV(c.A)
	} else {
		c.setCarryBit(c.Mem[v], 0)
		c.Mem[v] >>= 1
		c.setNV(c.Mem[v])
	}
}

func ROR(c *Cpu, b byte, v uint16, m Mode) {
	var s byte
	if c.C() {
		s = 0x80
	}
	if m == MODE_SNGL {
		c.A >>= 1
		c.A |= s
		c.setNV(c.A)
	} else {
		c.setCarryBit(c.Mem[v], 0)
		c.Mem[v] >>= 1
		c.Mem[v] |= s
		c.setNV(c.Mem[v])
	}
}

func BIT(c *Cpu, b byte, v uint16, m Mode) {
	if b&0x80 != 0 {
		c.P |= 0x80
	} else {
		c.P &= 0x7f
	}
	if b&0x40 != 0 {
		c.P |= 0x40
	} else {
		c.P &= 0xbf
	}
	if c.A&b != 0 {
		c.P &= 0xfd
	} else {
		c.P |= 0x02
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
	c.P = c.P&0xfe | b>>i
}

func EOR(c *Cpu, b byte, v uint16, m Mode) {
	c.A ^= b
	c.setNV(c.A)
}

func PHP(c *Cpu, b byte, v uint16, m Mode) {
	c.stackPush(c.P)
}

func PLP(c *Cpu, b byte, v uint16, m Mode) {
	c.P = c.stackPop()
}

func RTI(c *Cpu, b byte, v uint16, m Mode) {
	c.P = c.stackPop()
	c.PC = uint16(c.stackPop()) + uint16(c.stackPop())<<8
}

func TRB(c *Cpu, b byte, v uint16, m Mode) {
	if c.A&c.Mem[v] != 0 {
		c.P &= ^P_Z
	} else {
		c.P |= P_Z
	}
	c.Mem[v] &= ^c.A
}

func TSB(c *Cpu, b byte, v uint16, m Mode) {
	if c.A&c.Mem[v] != 0 {
		c.P &= ^P_Z
	} else {
		c.P |= P_Z
	}
	c.Mem[v] |= c.A
}

const null = 0

var Opcodes = []Instruction{
	/* F,  Imm,   ZP,  ZPX,  ZPY,  ABS, ABSX, ABSY,  IND, INDX, INDY, SNGL, BRA */
	{ADC, 0x69, 0x65, 0x75, null, 0x6d, 0x7d, 0x79, null, 0x61, 0x71, null, null},
	{AND, 0x29, 0x25, 0x35, null, 0x2d, 0x3d, 0x39, null, 0x21, 0x31, null, null},
	{ASL, null, 0x06, 0x16, null, 0x0e, 0x1e, null, null, null, null, 0x0a, null},
	{BCC, null, null, null, null, null, null, null, null, null, null, null, 0x90},
	{BCS, null, null, null, null, null, null, null, null, null, null, null, 0xb0},
	{BEQ, null, null, null, null, null, null, null, null, null, null, null, 0xf0},
	{BIT, null, 0x24, null, null, 0x2c, null, null, null, null, null, null, null},
	{BMI, null, null, null, null, null, null, null, null, null, null, null, 0x30},
	{BNE, null, null, null, null, null, null, null, null, null, null, null, 0xd0},
	{BPL, null, null, null, null, null, null, null, null, null, null, null, 0x10},
	{BRK, null, null, null, null, null, null, null, null, null, null, 0x00, null},
	{BVC, null, null, null, null, null, null, null, null, null, null, null, 0x50},
	{BVS, null, null, null, null, null, null, null, null, null, null, null, 0x70},
	{CLC, null, null, null, null, null, null, null, null, null, null, 0x18, null},
	{CLD, null, null, null, null, null, null, null, null, null, null, 0xd8, null},
	{CLI, null, null, null, null, null, null, null, null, null, null, 0x58, null},
	{CLV, null, null, null, null, null, null, null, null, null, null, 0xb8, null},
	{CMP, 0xc9, 0xc5, 0xd5, null, 0xcd, 0xdd, 0xd9, null, 0xc1, 0xd1, null, null},
	{CPX, 0xe0, 0xe4, null, null, 0xec, null, null, null, null, null, null, null},
	{CPY, 0xc0, 0xc4, null, null, 0xcc, null, null, null, null, null, null, null},
	{DEC, null, 0xc6, 0xd6, null, 0xce, 0xde, null, null, null, null, null, null},
	{DEX, null, null, null, null, null, null, null, null, null, null, 0xca, null},
	{DEY, null, null, null, null, null, null, null, null, null, null, 0x88, null},
	{EOR, 0x49, 0x45, 0x55, null, 0x4d, 0x5d, 0x59, null, 0x41, 0x51, null, null},
	{INC, null, 0xe6, 0xf6, null, 0xee, 0xfe, null, null, null, null, null, null},
	{INX, null, null, null, null, null, null, null, null, null, null, 0xe8, null},
	{INY, null, null, null, null, null, null, null, null, null, null, 0xc8, null},
	{JMP, null, null, null, null, 0x4c, null, null, 0x6c, null, null, null, null},
	{JSR, null, null, null, null, 0x20, null, null, null, null, null, null, null},
	{LDA, 0xa9, 0xa5, 0xb5, null, 0xad, 0xbd, 0xb9, null, 0xa1, 0xb1, null, null},
	{LDX, 0xa2, 0xa6, null, 0xb6, 0xae, null, 0xbe, null, null, null, null, null},
	{LDY, 0xa0, 0xa4, 0xb4, null, 0xac, 0xbc, null, null, null, null, null, null},
	{LSR, null, 0x46, 0x56, null, 0x4e, 0x5e, null, null, null, null, 0x4a, null},
	{NOP, null, null, null, null, null, null, null, null, null, null, 0xea, null},
	{ORA, 0x09, 0x05, 0x15, null, 0x0d, 0x1d, 0x19, null, 0x01, 0x11, null, null},
	{PHA, null, null, null, null, null, null, null, null, null, null, 0x48, null},
	{PHP, null, null, null, null, null, null, null, null, null, null, 0x08, null},
	{PLA, null, null, null, null, null, null, null, null, null, null, 0x68, null},
	{PLP, null, null, null, null, null, null, null, null, null, null, 0x28, null},
	{ROL, null, 0x26, 0x36, null, 0x2e, 0x3e, null, null, null, null, 0x2a, null},
	{ROR, null, 0x66, 0x76, null, 0x6e, 0x7e, null, null, null, null, 0x6a, null},
	{RTI, null, null, null, null, null, null, null, null, null, null, 0x40, null},
	{RTS, null, null, null, null, null, null, null, null, null, null, 0x60, null},
	{SBC, 0xe9, 0xe5, 0xf5, null, 0xed, 0xfd, 0xf9, null, 0xe1, 0xf1, null, null},
	{SEC, null, null, null, null, null, null, null, null, null, null, 0x38, null},
	{SED, null, null, null, null, null, null, null, null, null, null, 0xf8, null},
	{SEI, null, null, null, null, null, null, null, null, null, null, 0x78, null},
	{STA, null, 0x85, 0x95, null, 0x8d, 0x9d, 0x99, null, 0x81, 0x91, null, null},
	{STX, null, 0x86, null, 0x96, 0x8e, null, null, null, null, null, null, null},
	{STY, null, 0x84, 0x94, null, 0x8c, null, null, null, null, null, null, null},
	{TAX, null, null, null, null, null, null, null, null, null, null, 0xaa, null},
	{TAY, null, null, null, null, null, null, null, null, null, null, 0xa8, null},
	{TRB, null, 0x14, null, null, 0x1c, null, null, null, null, null, null, null},
	{TSB, null, 0x04, null, null, 0x0c, null, null, null, null, null, null, null},
	{TSX, null, null, null, null, null, null, null, null, null, null, 0xba, null},
	{TXA, null, null, null, null, null, null, null, null, null, null, 0x8a, null},
	{TXS, null, null, null, null, null, null, null, null, null, null, 0x9a, null},
	{TYA, null, null, null, null, null, null, null, null, null, null, 0x98, null},
}
