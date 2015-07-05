package gme

import (
	"io/ioutil"
	"testing"
)

func TestGME(t *testing.T) {
	b, err := ioutil.ReadFile("test.nsf")
	if err != nil {
		t.Fatal(err)
	}
	g, err := New(b, 44100)
	if err != nil {
		t.Fatal(err)
	}
	_, err = g.Track(0)
	if err != nil {
		t.Fatal(err)
	}
	if err := g.Start(0); err != nil {
		t.Fatal(err)
	}
	data := make([]int16, 4096)
	for i := 0; i < 10; i++ {
		if err := g.Play(data); err != nil {
			t.Error(err)
			break
		}
		if g.Ended() {
			t.Fatalf("ended")
		}
	}
	g.Close()
}
