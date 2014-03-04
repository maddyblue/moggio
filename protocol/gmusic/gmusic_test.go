package gmusic

import (
	"io/ioutil"
	"testing"
)

func TestGMusic(t *testing.T) {
	var err error
	var gm *GMusic
	pw, err := ioutil.ReadFile("pw")
	if err != nil {
		t.Fatal(err)
	}
	un := "matt.jibson"
	gm, err = Login(un, string(pw))
	if err != nil {
		t.Fatal(err)
	}
	_, err = gm.ListPlaylists()
	if err != nil {
		t.Fatal(err)
	}
	_, err = gm.ListPlaylistEntries()
	if err != nil {
		t.Fatal(err)
	}
	_, err = gm.ListTracks()
	if err != nil {
		t.Fatal(err)
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
