package main

import (
	"bytes"
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mjibson/mog/server"
	"gopkg.in/fsnotify.v1"

	// codecs
	_ "github.com/mjibson/mog/codec/mpa"
	_ "github.com/mjibson/mog/codec/nsf"

	// protocols
	"github.com/mjibson/mog/protocol/drive"
	"github.com/mjibson/mog/protocol/dropbox"
	_ "github.com/mjibson/mog/protocol/file"
	_ "github.com/mjibson/mog/protocol/gmusic"
	"github.com/mjibson/mog/protocol/soundcloud"
)

var (
	flagWatch      = flag.Bool("w", false, "watch current directory and exit on changes; for use with an autorestarter")
	flagDrive      = flag.String("drive", "", "Google Drive API credentials of the form ClientID:ClientSecret")
	flagDropbox    = flag.String("dropbox", "", "Dropbox API credentials of the form ClientID:ClientSecret")
	flagSoundcloud = flag.String("soundcloud", "", "SoundCloud API credentials of the form ClientID:ClientSecret")
)

func main() {
	flag.Parse()
	if *flagWatch {
		watch(".", "*.go", quit)
		base := filepath.Join("server", "static")
		src := filepath.Join(base, "src")
		watch(src, "*.js", jsx)
		go jsx()
	}
	redir := DefaultAddr
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
	log.Fatal(server.ListenAndServe("mog.state", DefaultAddr))
}

const DefaultAddr = ":6601"

func quit() {
	os.Exit(0)
}

func jsx() {
	base := filepath.Join("server", "static")
	src := filepath.Join(base, "src")
	js := filepath.Join(base, "js")
	log.Println("running jsx")
	var rs []io.Reader
	files := []string{
		"mog.js",

		"list.js",
		"protocol.js",

		"nav.js",
	}
	for _, name := range files {
		f, err := os.Open(filepath.Join(src, name))
		if err != nil {
			panic(err)
		}
		defer f.Close()
		rs = append(rs, f)
	}
	of, err := os.Create(filepath.Join(js, "mog.js"))
	if err != nil {
		log.Fatal(err)
	}
	defer of.Close()
	c := exec.Command("jsx")
	c.Stdin = io.MultiReader(rs...)
	c.Stdout = of
	buf := new(bytes.Buffer)
	c.Stderr = buf
	if err := c.Start(); err != nil {
		log.Fatal(err)
	}
	if err := c.Wait(); err != nil {
		log.Println(buf.String())
		log.Printf("jsx error: %v", err)
	}
	log.Println("jsx complete")
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
