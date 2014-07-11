package mog

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"testing"
	"time"

	_ "github.com/mjibson/mog/codec/mpa"
	_ "github.com/mjibson/mog/codec/nsf"
)

func TestServer(t *testing.T) {
	errs := make(chan error)
	go func() {
		errs <- ListenAndServe(DefaultAddr, "../codec")
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

	resp := fetch("/list", nil)
	b, _ := ioutil.ReadAll(resp.Body)
	songs := make(Songs)
	if err := json.Unmarshal(b, &songs); err != nil {
		t.Fatal(err)
	}
	v := make(url.Values)
	for i, _ := range songs {
		if i < 10 {
			v.Add("add", strconv.Itoa(i))
		}
	}
	if len(v) == 0 {
		t.Fatal("expected songs")
	}
	resp = fetch("/playlist/change", v)
	b, _ = ioutil.ReadAll(resp.Body)
	var pc PlaylistChange
	if err := json.Unmarshal(b, &pc); err != nil {
		t.Fatal(err)
	}
	resp = fetch("/playlist/get", nil)
	b, _ = ioutil.ReadAll(resp.Body)
	var pl Playlist
	if err := json.Unmarshal(b, &pl); err != nil {
		t.Fatal(err)
	}
	resp = fetch("/play", nil)
	select {}
}
