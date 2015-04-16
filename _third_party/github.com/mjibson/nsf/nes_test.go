package nsf

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/mjibson/mog/_third_party/github.com/mjibson/nsf/cpu6502"
)

func loadNES(fname string) *NSF {
	var err error
	var n NSF
	b, err := ioutil.ReadFile(fname)
	if err != nil {
		panic(err)
	}
	if string(b[:4]) != "NES\u001a" {
		panic("not a NES file")
	}
	prg := b[4]
	//chr := n.b[5]
	n.Data = b[16:]
	mapper := b[6]>>4 | b[7]&0xF0
	if mapper != 0 {
		panic("unknown mapper")
	}
	n.ram = new(ram)
	n.Cpu = cpu6502.New(n.ram)
Loop:
	for a := 0x4000; true; {
		for i := 0; i < int(prg); i++ {
			a += 0x4000
			if a > 0xffff {
				break Loop
			}
			copy(n.ram.M[a:a+0x4000], n.Data[i*0x4000:(i+1)*0x4000])
		}
	}
	n.Cpu.Reset()
	if n.Cpu.PC == 0 {
		panic("PC == 0")
	}
	return &n
}

func TestNesTest(t *testing.T) {
	f, _ := os.Open("roms/nestest/nestest.log")
	s := bufio.NewScanner(f)
	n := loadNES("roms/nestest/nestest.nes")
	n.Cpu.L = make([]cpu6502.Log, 10)
	i := 0
	n.Cpu.PC = 0xC000
	defer func() {
		t.Log("instructions", i)
		t.Log(strings.Fields(s.Text()))
		t.Log(n.Cpu.StringLog())
		t.Log(n.Cpu)
	}()
	for {
		i++
		if !s.Scan() {
			if i == 8992 {
				return
			}
			t.Fatal("expected scan")
		} else if s.Err() != nil {
			t.Fatal(s.Err())
		}
		l := s.Text()
		if l[0:4] != fmt.Sprintf("%04X", n.Cpu.PC) {
			t.Fatal("bad pc")
		}
		if l[6:8] != fmt.Sprintf("%02X", n.ram.Read(n.Cpu.PC)) {
			t.Fatal("bad i")
		}
		if l[50:52] != fmt.Sprintf("%02X", n.Cpu.A) {
			t.Fatal("bad a")
		}
		if l[55:57] != fmt.Sprintf("%02X", n.Cpu.X) {
			t.Fatal("bad x")
		}
		if l[60:62] != fmt.Sprintf("%02X", n.Cpu.Y) {
			t.Fatal("bad y")
		}
		if l[65:67] != fmt.Sprintf("%02X", n.Cpu.P) {
			t.Fatal("bad p")
		}
		if l[71:73] != fmt.Sprintf("%02X", n.Cpu.S) {
			t.Fatal("bad s")
		}
		n.Cpu.Step()
	}
}
