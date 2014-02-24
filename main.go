package main

import (
	"log"

	"github.com/mjibson/mog/mog"
)

func main() {
	log.Fatal(mog.ListenAndServe(":6601", "."))
}
