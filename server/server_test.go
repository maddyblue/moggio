package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"net/http"
	"net/url"
	"testing"
	"time"

	_ "github.com/mjibson/mog/codec/mpa"
	_ "github.com/mjibson/mog/codec/nsf"
	_ "github.com/mjibson/mog/protocol/file"
	_ "github.com/mjibson/mog/protocol/gmusic"
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
			if resp.StatusCode != 200 {
				b, _ := ioutil.ReadAll(resp.Body)
				errs <- fmt.Errorf("%s: %s", resp.Status, string(b))
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
	resp = fetch("/api/protocol/update", url.Values{
		"protocol": []string{"file"},
		"params":   []string{".."},
	})
	resp = fetch("/api/list", nil)
	if resp.StatusCode != 200 {
		t.Fatal("bad status")
	}
	songs := make([]string, 0)
	if err := json.NewDecoder(resp.Body).Decode(&songs); err != nil {
		t.Fatal(err)
	}
	if len(songs) == 0 {
		t.Fatal("expected songs")
	}
	v := url.Values{
		"add": []string{
			"file|1-../mm3.nsf",
			"file|2-../mm3.nsf",
			"file|3-../mm3.nsf",
			"file|4-../mm3.nsf",
		},
	}
	resp = fetch("/api/playlist/change", v)
	var pc PlaylistChange
	if err := json.NewDecoder(resp.Body).Decode(&pc); err != nil {
		t.Fatal(err)
	}
	resp = fetch("/api/cmd/play", nil)
	time.Sleep(time.Second * 2)
	resp = fetch("/api/cmd/next", nil)
	time.Sleep(time.Second * 2)
	resp = fetch("/api/cmd/next", nil)
	time.Sleep(time.Second * 2)
	resp = fetch("/api/cmd/next", nil)
	time.Sleep(time.Second * 2)
}
