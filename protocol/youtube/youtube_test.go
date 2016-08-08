package youtube

import (
	"fmt"
	"testing"
)

func TestYoutube(t *testing.T) {
	p, err := New([]string{"https://www.youtube.com/watch?v=wZNYDzNGB-Q"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	_ = p
	s, err := p.GetSong("")
	if err != nil {
		t.Fatal(err)
	}
	sr, ch, err := s.Init()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("SR", sr, "CH", ch)
	samp, err := s.Play(256)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(samp)
}
