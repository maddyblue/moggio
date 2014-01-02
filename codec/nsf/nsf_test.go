package nsf

import (
	"os"
	"testing"
)

func TestNSF(t *testing.T) {
	f, err := os.Open("mm3.nsf")
	if err != nil {
		t.Fatal(err)
	}
	n, err := ReadNSF(f)
	if err != nil {
		t.Fatal(err)
	}
	if n.Load != 0x8000 || n.Init != 0x8003 || n.Play != 0x8000 {
		t.Error("bad addresses")
	}
}
