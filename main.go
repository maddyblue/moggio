package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mjibson/mog/_third_party/github.com/facebookgo/httpcontrol"
	"github.com/mjibson/mog/_third_party/gopkg.in/fsnotify.v1"
	"github.com/mjibson/mog/server"

	// codecs
	_ "github.com/mjibson/mog/codec/flac"
	_ "github.com/mjibson/mog/codec/gme"
	_ "github.com/mjibson/mog/codec/mpa"
	_ "github.com/mjibson/mog/codec/nsf"
	_ "github.com/mjibson/mog/codec/vorbis"
	_ "github.com/mjibson/mog/codec/wav"

	// protocols
	_ "github.com/mjibson/mog/protocol/bandcamp"
	"github.com/mjibson/mog/protocol/drive"
	"github.com/mjibson/mog/protocol/dropbox"
	_ "github.com/mjibson/mog/protocol/file"
	_ "github.com/mjibson/mog/protocol/gmusic"
	"github.com/mjibson/mog/protocol/soundcloud"
	_ "github.com/mjibson/mog/protocol/stream"
)

var (
	flagAddr       = flag.String("addr", ":6601", "listen address")
	flagWatch      = flag.Bool("w", false, "watch current directory and exit on changes; for use with an autorestarter")
	flagDrive      = flag.String("drive", "792434736327-0pup5skbua0gbfld4min3nfv2reairte.apps.googleusercontent.com:OsN_bydWG45resaU0PPiDmtK", "Google Drive API credentials of the form ClientID:ClientSecret")
	flagDropbox    = flag.String("dropbox", "rnhpqsbed2q2ezn:ldref688unj74ld", "Dropbox API credentials of the form ClientID:ClientSecret")
	flagSoundcloud = flag.String("soundcloud", "ec28c2226a0838d01edc6ed0014e462e:a115e94029d698f541960c8dc8560978", "SoundCloud API credentials of the form ClientID:ClientSecret")
	flagDev        = flag.Bool("dev", false, "enable dev mode")
	stateFile      = flag.String("state", "", "specify non-default statefile location")
)

func main() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
	http.DefaultClient = &http.Client{
		Transport: &httpcontrol.Transport{
			ResponseHeaderTimeout: time.Second * 3,
			MaxTries:              3,
			RetryAfterTimeout:     true,
		},
	}
	if *flagWatch {
		watch(".", "*.go", quit)
		go browserify()
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
			*stateFile = "mog.state"
		case runtime.GOOS == "windows":
			dir := filepath.Join(os.Getenv("APPDATA"), "mog")
			if err := os.MkdirAll(dir, 0600); err != nil {
				log.Fatal(err)
			}
			*stateFile = filepath.Join(dir, "mog.state")
		default:
			*stateFile = filepath.Join(os.Getenv("HOME"), ".mog.state")
		}
	}
	log.Fatal(server.ListenAndServe(*stateFile, *flagAddr, *flagDev))
}

func quit() {
	os.Exit(0)
}

func browserify() {
	base := filepath.Join("server", "static")
	src := filepath.Join(base, "src")
	js := filepath.Join(base, "js")
	log.Println("starting watchify")
	c := exec.Command("watchify",
		"-t", "[", "reactify", "--es6", "]",
		filepath.Join(src, "nav.js"),
		"-o", filepath.Join(js, "mog.js"),
		"--verbose",
	)
	c.Stderr = os.Stderr
	c.Stdout = os.Stdout
	if err := c.Start(); err != nil {
		log.Fatal(err)
	}
	if err := c.Wait(); err != nil {
		log.Printf("browserify error: %v", err)
	}
}

func run(name string, arg ...string) func() {
	return func() {
		log.Println("running", name)
		c := exec.Command(name, arg...)
		stdout, err := c.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		stderr, err := c.StderrPipe()
		if err != nil {
			log.Fatal(err)
		}
		if err := c.Start(); err != nil {
			log.Fatal(err)
		}
		go func() { io.Copy(os.Stdout, stdout) }()
		go func() { io.Copy(os.Stderr, stderr) }()
		if err := c.Wait(); err != nil {
			log.Printf("run error: %v: %v", name, err)
		}
		log.Println("run complete:", name)
	}
}

func watch(root, pattern string, f func()) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if matched, err := filepath.Match(pattern, info.Name()); err != nil {
			log.Fatal(err)
		} else if !matched {
			return nil
		}
		err = watcher.Add(path)
		if err != nil {
			log.Fatal(err)
		}
		return nil
	})
	log.Println("watching", pattern, "in", root)
	wait := time.Now()
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if wait.After(time.Now()) {
					continue
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					f()
					wait = time.Now().Add(time.Second * 2)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()
}

//go:generate esc -o server/static.go -pkg server -prefix server server/static/index.html server/static/css server/static/fonts server/static/js
