package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/howeyc/fsnotify"
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
		watch()
	}
	log.Fatal(server.ListenAndServe(":6601"))
}

func watch() {
	time.Sleep(time.Second)
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Print(err)
		return
	}
	go func() {
		for {
			select {
			case ev := <-watcher.Event:
				log.Print("file changed, exiting:", ev)
				os.Exit(0)
			case err := <-watcher.Error:
				log.Print(err)
			}
		}
	}()
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() || (len(path) > 1 && path[0] == '.') {
			return nil
		}
		watcher.Watch(path)
		return nil
	})
}
