package gmusic

import (
	"io/ioutil"
	"testing"
)

func TestGMusic(t *testing.T) {
	username, err := ioutil.ReadFile("username")
	if err != nil {
		t.Fatal(err)
	}
	password, err := ioutil.ReadFile("password")
	if err != nil {
		t.Fatal(err)
	}
	gm, err := Login(string(username), string(password))
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
	resp, err := gm.GetStream("51a33f68-390a-3f8a-b4e1-4a2e8d82df65")
	if err != nil {
		t.Fatal(err)
	}
	b := make([]byte, 2)
	resp.Body.Read(b)
	if b[0] != 0xff || b[1] != 0xfb {
		t.Fatal("expected mp3 header, got:", b)
	}
}
