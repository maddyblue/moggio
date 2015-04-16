package nsf

import (
	"os"
	"testing"

	"github.com/mjibson/mog/output"
)

func TestNsf(t *testing.T) {
	f, err := os.Open("mm3.nsf")
	if err != nil {
		t.Fatal(err)
	}
	n, err := New(f)
	if err != nil {
		t.Fatal(err)
	}
	if n.LoadAddr != 0x8000 || n.InitAddr != 0x8003 || n.PlayAddr != 0x8000 {
		t.Fatal("bad addresses")
	}
	n.Init(1)
	o, err := output.Get(int(n.SampleRate), 1)
	if err != nil {
		t.Fatal(err)
	}
	const div = 10
	ns := int(n.SampleRate / div)
	for {
		o.Push(n.Play(ns))
	}
}

func TestNsfe(t *testing.T) {
	f, err := os.Open("mm3.nsfe")
	if err != nil {
		t.Fatal(err)
	}
	n, err := New(f)
	if err != nil {
		t.Fatal(err)
	}
	if n.LoadAddr != 0x8000 || n.InitAddr != 0x8003 || n.PlayAddr != 0x8000 {
		t.Fatal("bad addresses")
	}
	n.Init(1)
	o, err := output.Get(int(n.SampleRate), 1)
	if err != nil {
		t.Fatal(err)
	}
	const div = 10
	ns := int(n.SampleRate / div)
	for {
		o.Push(n.Play(ns))
	}
}
