package mog

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	errs := make(chan error)
	go func() {
		errs <- ListenAndServe(DefaultAddr, "../codec/nsf")
	}()

	fetch := func(path string) *http.Response {
		rc := make(chan *http.Response)
		go func() {
			u := &url.URL{
				Scheme: "http",
				Host:   DefaultAddr,
				Path:   path,
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

	resp := fetch("/list")
	b, _ := ioutil.ReadAll(resp.Body)
	t.Fatal(string(b))
}
