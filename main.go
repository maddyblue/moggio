package main

import (
	"flag"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/go-fsnotify/fsnotify"
	_ "github.com/mjibson/mog/codec/mpa"
	_ "github.com/mjibson/mog/codec/nsf"
	_ "github.com/mjibson/mog/protocol/file"
	_ "github.com/mjibson/mog/protocol/gmusic"
	"github.com/mjibson/mog/server"
)

var (
	flagWatch = flag.Bool("w", false, "watch current directory and exit on changes; for use with an autorestarter")
)

func main() {
	flag.Parse()
	if *flagWatch {
		watch(".", "*.go", quit)
		jsx := run("jsx", "server/static/src", "server/static/scripts")
		watch("server/static/src", "*.js", jsx)
		jsx()
	}
	log.Fatal(server.ListenAndServe(":6601"))
}

func quit() {
	os.Exit(0)
}

func run(name string, arg ...string) func() {
	return func() {
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
