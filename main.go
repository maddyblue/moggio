package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/facebookgo/httpcontrol"
	"github.com/mjibson/moggio/server"

	// codecs
	_ "github.com/mjibson/moggio/codec/aac"
	_ "github.com/mjibson/moggio/codec/flac"
	_ "github.com/mjibson/moggio/codec/gme"
	_ "github.com/mjibson/moggio/codec/mpa"
	_ "github.com/mjibson/moggio/codec/nsf"
	_ "github.com/mjibson/moggio/codec/rar"
	_ "github.com/mjibson/moggio/codec/vorbis"
	_ "github.com/mjibson/moggio/codec/wav"

	// protocols
	_ "github.com/mjibson/moggio/protocol/bandcamp"
	"github.com/mjibson/moggio/protocol/drive"
	"github.com/mjibson/moggio/protocol/dropbox"
	_ "github.com/mjibson/moggio/protocol/file"
	_ "github.com/mjibson/moggio/protocol/gmusic"
	"github.com/mjibson/moggio/protocol/soundcloud"
	_ "github.com/mjibson/moggio/protocol/stream"
	_ "github.com/mjibson/moggio/protocol/youtube"
)

var (
	flagAddr       = flag.String("addr", ":6601", "listen address")
	flagDrive      = flag.String("drive", "792434736327-0pup5skbua0gbfld4min3nfv2reairte.apps.googleusercontent.com:OsN_bydWG45resaU0PPiDmtK", "Google Drive API credentials of the form ClientID:ClientSecret")
	flagDropbox    = flag.String("dropbox", "rnhpqsbed2q2ezn:ldref688unj74ld", "Dropbox API credentials of the form ClientID:ClientSecret")
	flagSoundcloud = flag.String("soundcloud", "ec28c2226a0838d01edc6ed0014e462e:a115e94029d698f541960c8dc8560978", "SoundCloud API credentials of the form ClientID:ClientSecret")
	flagDev        = flag.Bool("dev", false, "enable dev mode")
	//flagCentral = flag.String("central", "https://moggio-music-client.appspot.com", "Central Moggio data server; empty to disable")
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
	redir := *flagAddr
	if strings.HasPrefix(redir, ":") {
		redir = "localhost" + redir
	}
	redir = "http://" + redir + "/api/oauth/"
	if *flagDrive != "" {
		sp := strings.Split(*flagDrive, ":")
		if len(sp) != 2 {
			log.Fatal("bad drive string %s", *flagDrive)
		}
		drive.Init(sp[0], sp[1], redir)
	}
	if *flagDropbox != "" {
		sp := strings.Split(*flagDropbox, ":")
		if len(sp) != 2 {
			log.Fatal("bad drive string %s", *flagDropbox)
		}
		dropbox.Init(sp[0], sp[1], redir)
	}
	if *flagSoundcloud != "" {
		sp := strings.Split(*flagSoundcloud, ":")
		if len(sp) != 2 {
			log.Fatal("bad drive string %s", *flagSoundcloud)
		}
		soundcloud.Init(sp[0], sp[1], redir)
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
	log.Fatal(server.ListenAndServe(*stateFile, *flagAddr, "", *flagDev))
}

//go:generate browserify -t [ reactify --es6 ] server/static/src/nav.js -o server/static/js/moggio.js
//go:generate esc -o server/static.go -pkg server -prefix server server/static/index.html server/static/css server/static/fonts server/static/js
