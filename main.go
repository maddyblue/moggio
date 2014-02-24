package main

import (
	"log"

	_ "github.com/mjibson/mog/codec/nsf"
	"github.com/mjibson/mog/mog"
)

func main() {
	log.Fatal(mog.ListenAndServe(":6601", "."))
}
