package cpu6502

import "testing"

type CpuTest struct {
	Mem []byte
	End Cpu
}

var CpuTests = []CpuTest{
	// load, set
	CpuTest{
		Mem: []byte{0xa9, 0x01, 0x8d, 0x00, 0x02, 0xa9, 0x05, 0x8d, 0x01, 0x02, 0xa9, 0x08, 0x8d, 0x02, 0x02},
		End: Cpu{
			A:  0x08,
			S:  0xff,
			PC: 0x0610,
			P:  0x30,
		},
	},
	// load, transfer, increment, add
	CpuTest{
		Mem: []byte{0xa9, 0xc0, 0xaa, 0xe8, 0x69, 0xc4, 0x00},
		End: Cpu{
			A:  0x84,
			X:  0xc1,
			S:  0xff,
			PC: 0x0607,
			P:  0xb1,
		},
	},
	// bne
	CpuTest{
		Mem: []byte{0xa2, 0x08, 0xca, 0x8e, 0x00, 0x02, 0xe0, 0x03, 0xd0, 0xf8, 0x8e, 0x01, 0x02, 0x00},
		End: Cpu{
			X:  0x03,
			S:  0xff,
			PC: 0x060e,
			P:  0x33,
		},
	},
}

func Test6502(t *testing.T) {
	for _, test := range CpuTests {
		c := New()
		copy(c.Mem[c.PC:], test.Mem)
		c.Run()
		if c.A != test.End.A ||
			c.X != test.End.X ||
			c.Y != test.End.Y ||
			c.S != test.End.S ||
			c.PC != test.End.PC ||
			c.P != test.End.P {
			t.Fatalf("bad cpu state, got:\n%sexpected:\n%s", c, &test.End)
		}
	}
}
