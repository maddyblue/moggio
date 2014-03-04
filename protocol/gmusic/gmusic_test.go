package gmusic

import (
	"io/ioutil"
	"log"
	"testing"
)

func TestGMusic(t *testing.T) {
	pw, err := ioutil.ReadFile("pw")
	if err != nil {
		t.Fatal(err)
	}
	un := "matt.jibson"
	gm, err := Login(un, string(pw))
	if err != nil {
		t.Fatal(err)
	}
	playlists, err := gm.ListPlaylists()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range playlists {
		log.Println(p.Name)
	}
}

func _TestBadLogin(t *testing.T) {
	pw := "onethsochk"
	un := "blah"
	_, err := Login(un, pw)
	if err == nil {
		t.Fatal("expected error")
	}
}
