package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/facebookgo/httpcontrol"
	"github.com/mjibson/moggio/server"

	// codecs
	_ "github.com/mjibson/moggio/codec/flac"
	_ "github.com/mjibson/moggio/codec/gme"
	_ "github.com/mjibson/moggio/codec/mpa"
	_ "github.com/mjibson/moggio/codec/nsf"
	_ "github.com/mjibson/moggio/codec/rar"
	_ "github.com/mjibson/moggio/codec/vorbis"
	_ "github.com/mjibson/moggio/codec/wav"

	// protocols
	_ "github.com/mjibson/moggio/protocol/file"
	_ "github.com/mjibson/moggio/protocol/stream"
)

var (
	flagAddr  = flag.String("addr", ":6601", "listen address")
	flagDev   = flag.Bool("dev", false, "enable dev mode")
	stateFile = flag.String("state", "", "specify non-default statefile location")
)

func main() {
	flag.Parse()
	http.DefaultClient = &http.Client{
		Transport: &httpcontrol.Transport{
			ResponseHeaderTimeout: time.Second * 3,
			MaxTries:              3,
			RetryAfterTimeout:     true,
		},
	}
	if *stateFile == "" {
		switch {
		case *flagDev:
			*stateFile = "moggio.state"
		case runtime.GOOS == "windows":
			dir := filepath.Join(os.Getenv("APPDATA"), "moggio")
			if err := os.MkdirAll(dir, 0600); err != nil {
				log.Fatal(err)
			}
			*stateFile = filepath.Join(dir, "moggio.state")
		default:
			*stateFile = filepath.Join(os.Getenv("HOME"), ".moggio.state")
		}
	}
	log.Fatal(server.ListenAndServe(*stateFile, *flagAddr, *flagDev))
}

//go:generate browserify -t [ reactify --es6 ] server/static/src/nav.js -o server/static/js/moggio.js
