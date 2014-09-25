package server

import (
	"encoding/json"
	"io/ioutil"

	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	_ "github.com/mjibson/mog/codec/mpa"
	_ "github.com/mjibson/mog/codec/nsf"
	_ "github.com/mjibson/mog/protocol/file"
	_ "github.com/mjibson/mog/protocol/gmusic"
)

func TestServer(t *testing.T) {
	srv, err := New("")
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.GetMux())
	defer ts.Close()
	client := &http.Client{Timeout: time.Second}
	fetch := func(path string, values url.Values) *http.Response {
		u := ts.URL + path + "?" + values.Encode()
		t.Log("fetching", u)
		resp, err := client.Get(u)
		if err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200 {
			b, _ := ioutil.ReadAll(resp.Body)
			t.Fatalf("%s: %s", resp.Status, string(b))
		}
		return resp
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
	time.Sleep(time.Second)
	resp = fetch("/api/cmd/next", nil)
	time.Sleep(time.Second)
	resp = fetch("/api/cmd/next", nil)
	time.Sleep(time.Second)
	resp = fetch("/api/cmd/pause", nil)
	time.Sleep(time.Second)
	resp = fetch("/api/cmd/pause", nil)
	time.Sleep(time.Second)
	resp = fetch("/api/cmd/prev", nil)
	time.Sleep(time.Second)
	resp = fetch("/api/cmd/prev", nil)
	time.Sleep(time.Second)
}
