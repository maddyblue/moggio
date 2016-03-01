package server

import (
	"bytes"
	"encoding/json"
	"io/ioutil"

	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	_ "github.com/mjibson/moggio/codec/mpa"
	_ "github.com/mjibson/moggio/codec/nsf"
	_ "github.com/mjibson/moggio/protocol/file"
)

func TestServer(t *testing.T) {
	srv, err := New("")
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(srv.GetMux(true))
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
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return resp
	}

	var resp *http.Response
	resp = fetch("/api/protocol/add", url.Values{
		"protocol": []string{"file"},
		"params":   []string{"../m"},
	})
	resp = fetch("/api/list", nil)
	if resp.StatusCode != 200 {
		t.Fatal("bad status")
	}
	songs := make([]SongID, 0)
	if err := json.NewDecoder(resp.Body).Decode(&songs); err != nil {
		t.Fatal(err)
	}
	if len(songs) == 0 {
		t.Fatal("expected songs")
	}
	v := url.Values{
		"add": []string{
			"file|../m|1-../m/mm3.nsf",
			"file|../m|2-../m/mm3.nsf",
			"file|../m|3-../m/mm3.nsf",
			"file|../m|4-../m/mm3.nsf",
		},
	}
	resp = fetch("/api/playlist/change", v)
	var pc PlaylistChange
	if err := json.NewDecoder(resp.Body).Decode(&pc); err != nil {
		t.Fatal(err)
	}
	if len(pc.Errors) > 0 {
		t.Fatal(pc.Errors[0])
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
