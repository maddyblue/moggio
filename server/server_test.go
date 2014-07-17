package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	_ "github.com/mjibson/mog/codec/mpa"
	_ "github.com/mjibson/mog/codec/nsf"
	_ "github.com/mjibson/mog/protocol/file"
)

func TestServer(t *testing.T) {
	errs := make(chan error)
	go func() {
		errs <- ListenAndServe(DefaultAddr)
	}()
	time.Sleep(time.Millisecond * 100)
	fetch := func(path string, values url.Values) *http.Response {
		rc := make(chan *http.Response)
		go func() {
			u := &url.URL{
				Scheme:   "http",
				Host:     DefaultAddr,
				Path:     path,
				RawQuery: values.Encode(),
			}
			t.Log("fetching", u)
			resp, err := http.Get(u.String())
			if err != nil {
				errs <- err
				return
			}
			rc <- resp
		}()
		select {
		case <-time.After(time.Second):
			t.Fatal("timeout")
		case err := <-errs:
			t.Fatal(err)
		case resp := <-rc:
			return resp
		}
		panic("unreachable")
	}

	var resp *http.Response
	resp = fetch("/protocol/update", url.Values{
		"protocol": []string{"file"},
		"params":   []string{".."},
	})
	resp = fetch("/list", nil)
	if resp.StatusCode != 200 {
		t.Fatal("bad status")
	}
	songs := make([][]string, 0)
	if err := json.NewDecoder(resp.Body).Decode(&songs); err != nil {
		t.Fatal(err)
	}
	v := make(url.Values)
	for i, s := range songs {
		if i < 10 {
			v.Add("add", fmt.Sprintf("%v|%v", s[0], s[1]))
		}
	}
	if len(v) == 0 {
		t.Fatal("expected songs")
	}
	resp = fetch("/playlist/change", v)
	var pc PlaylistChange
	if err := json.NewDecoder(resp.Body).Decode(&pc); err != nil {
		t.Fatal(err)
	}
	resp = fetch("/playlist/get", nil)
	var pl Playlist
	if err := json.NewDecoder(resp.Body).Decode(&pl); err != nil {
		t.Fatal(err)
	}
	resp = fetch("/play", nil)
	select {}
}
