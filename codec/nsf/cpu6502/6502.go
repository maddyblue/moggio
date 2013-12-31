package cpu6502

import "fmt"

type opcode byte

type Instruction struct {
	Name            string
	Imm             opcode
	ZP, ZPX, ZPY    opcode
	ABS, ABSX, ABSY opcode
	IND, INDX, INDY opcode
	SNGL, BRA       opcode
}

var Optable map[opcode]*Op

type Op struct {
	F    func(*Cpu, byte)
	Name string
	Mode mode
}

type mode int

const (
	MODE_IMM mode = iota
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
}

func (c *Cpu) Step() {
	inst := c.Mem[c.PC]
	c.PC++
	o, ok := Optable[opcode(inst)]
	if !ok {
		return
	}
	var v byte
	switch o.Mode {
	case MODE_IMM:
		v = c.Mem[c.PC]
		c.PC++
	default:
		panic("6502: bad address mode")
	}
	o.F(c, v)
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
func (c *Cpu) CLC() { c.P &= 0xfe }
func (c *Cpu) SEV() { c.P |= P_V }
func (c *Cpu) CLV() { c.P &= 0xbf }

func (c *Cpu) C() bool       { return c.p(P_C) }
func (c *Cpu) V() bool       { return c.p(P_V) }
func (c *Cpu) p(v byte) bool { return c.P&v != 0 }

const (
	P_C = 0x01
	P_V = 0x40
)

func (c *Cpu) Print() {
	const f = "%2s: %5d 0x%04[2]X %016[2]b\n"
	fmt.Printf(f, "A", c.A)
	fmt.Printf(f, "X", c.X)
	fmt.Printf(f, "Y", c.Y)
	fmt.Printf(f, "P", c.P)
	fmt.Printf(f, "PC", c.PC)
}

func init() {
	Optable = make(map[opcode]*Op)
	for _, inst := range Opcodes {
		if inst.Imm != null {
			Optable[inst.Imm] = &Op{
				F:    Functions[inst.Name],
				Name: inst.Name,
				Mode: MODE_IMM,
			}
		}
	}
}

var Functions = map[string]func(*Cpu, byte){
	"ADC": func(c *Cpu, v byte) {
		if (c.A^v)&0x80 != 0 {
			c.CLV()
		} else {
			c.SEV()
		}
		a := int(c.A) + int(v)
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
	},
}

const null = 0

var Opcodes = []Instruction{
	/* Name, Imm,  ZP,   ZPX,  ZPY,  ABS, ABSX, ABSY,  IND, INDX, INDY, SNGL, BRA */
	{"ADC", 0x69, 0x65, 0x75, null, 0x6d, 0x7d, 0x79, null, 0x61, 0x71, null, null},
	{"AND", 0x29, 0x25, 0x35, null, 0x2d, 0x3d, 0x39, null, 0x21, 0x31, null, null},
	{"ASL", null, 0x06, 0x16, null, 0x0e, 0x1e, null, null, null, null, 0x0a, null},
	{"BCC", null, null, null, null, null, null, null, null, null, null, null, 0x90},
	{"BCS", null, null, null, null, null, null, null, null, null, null, null, 0xb0},
	{"BEQ", null, null, null, null, null, null, null, null, null, null, null, 0xf0},
	{"BIT", null, 0x24, null, null, 0x2c, null, null, null, null, null, null, null},
	{"BMI", null, null, null, null, null, null, null, null, null, null, null, 0x30},
	{"BNE", null, null, null, null, null, null, null, null, null, null, null, 0xd0},
	{"BPL", null, null, null, null, null, null, null, null, null, null, null, 0x10},
	{"BRK", null, null, null, null, null, null, null, null, null, null, 0x00, null},
	{"BVC", null, null, null, null, null, null, null, null, null, null, null, 0x50},
	{"BVS", null, null, null, null, null, null, null, null, null, null, null, 0x70},
	{"CLC", null, null, null, null, null, null, null, null, null, null, 0x18, null},
	{"CLD", null, null, null, null, null, null, null, null, null, null, 0xd8, null},
	{"CLI", null, null, null, null, null, null, null, null, null, null, 0x58, null},
	{"CLV", null, null, null, null, null, null, null, null, null, null, 0xb8, null},
	{"CMP", 0xc9, 0xc5, 0xd5, null, 0xcd, 0xdd, 0xd9, null, 0xc1, 0xd1, null, null},
	{"CPX", 0xe0, 0xe4, null, null, 0xec, null, null, null, null, null, null, null},
	{"CPY", 0xc0, 0xc4, null, null, 0xcc, null, null, null, null, null, null, null},
	{"DEC", null, 0xc6, 0xd6, null, 0xce, 0xde, null, null, null, null, null, null},
	{"DEX", null, null, null, null, null, null, null, null, null, null, 0xca, null},
	{"DEY", null, null, null, null, null, null, null, null, null, null, 0x88, null},
	{"EOR", 0x49, 0x45, 0x55, null, 0x4d, 0x5d, 0x59, null, 0x41, 0x51, null, null},
	{"INC", null, 0xe6, 0xf6, null, 0xee, 0xfe, null, null, null, null, null, null},
	{"INX", null, null, null, null, null, null, null, null, null, null, 0xe8, null},
	{"INY", null, null, null, null, null, null, null, null, null, null, 0xc8, null},
	{"JMP", null, null, null, null, 0x4c, null, null, 0x6c, null, null, null, null},
	{"JSR", null, null, null, null, 0x20, null, null, null, null, null, null, null},
	{"LDA", 0xa9, 0xa5, 0xb5, null, 0xad, 0xbd, 0xb9, null, 0xa1, 0xb1, null, null},
	{"LDX", 0xa2, 0xa6, null, 0xb6, 0xae, null, 0xbe, null, null, null, null, null},
	{"LDY", 0xa0, 0xa4, 0xb4, null, 0xac, 0xbc, null, null, null, null, null, null},
	{"LSR", null, 0x46, 0x56, null, 0x4e, 0x5e, null, null, null, null, 0x4a, null},
	{"NOP", null, null, null, null, null, null, null, null, null, null, 0xea, null},
	{"ORA", 0x09, 0x05, 0x15, null, 0x0d, 0x1d, 0x19, null, 0x01, 0x11, null, null},
	{"PHA", null, null, null, null, null, null, null, null, null, null, 0x48, null},
	{"PHP", null, null, null, null, null, null, null, null, null, null, 0x08, null},
	{"PLA", null, null, null, null, null, null, null, null, null, null, 0x68, null},
	{"PLP", null, null, null, null, null, null, null, null, null, null, 0x28, null},
	{"ROL", null, 0x26, 0x36, null, 0x2e, 0x3e, null, null, null, null, 0x2a, null},
	{"ROR", null, 0x66, 0x76, null, 0x6e, 0x7e, null, null, null, null, 0x6a, null},
	{"RTI", null, null, null, null, null, null, null, null, null, null, 0x40, null},
	{"RTS", null, null, null, null, null, null, null, null, null, null, 0x60, null},
	{"SBC", 0xe9, 0xe5, 0xf5, null, 0xed, 0xfd, 0xf9, null, 0xe1, 0xf1, null, null},
	{"SEC", null, null, null, null, null, null, null, null, null, null, 0x38, null},
	{"SED", null, null, null, null, null, null, null, null, null, null, 0xf8, null},
	{"SEI", null, null, null, null, null, null, null, null, null, null, 0x78, null},
	{"STA", null, 0x85, 0x95, null, 0x8d, 0x9d, 0x99, null, 0x81, 0x91, null, null},
	{"STX", null, 0x86, null, 0x96, 0x8e, null, null, null, null, null, null, null},
	{"STY", null, 0x84, 0x94, null, 0x8c, null, null, null, null, null, null, null},
	{"TAX", null, null, null, null, null, null, null, null, null, null, 0xaa, null},
	{"TAY", null, null, null, null, null, null, null, null, null, null, 0xa8, null},
	{"TSX", null, null, null, null, null, null, null, null, null, null, 0xba, null},
	{"TXA", null, null, null, null, null, null, null, null, null, null, 0x8a, null},
	{"TXS", null, null, null, null, null, null, null, null, null, null, 0x9a, null},
	{"TYA", null, null, null, null, null, null, null, null, null, null, 0x98, null},
}
