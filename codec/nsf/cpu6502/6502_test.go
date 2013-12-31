package cpu6502

import "testing"

func Test6502(t *testing.T) {
	c := Cpu{}
	copy(c.Mem[:], []byte{
		0x69, 0x01,
	})
	c.Step()
	c.Print()
}
