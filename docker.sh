#!/bin/sh

set -e

build()
{
	export GOOS=$1
	export GOARCH=$2
	EXT=""
	if [ $GOOS = "windows" ]; then
		EXT=".exe"
	fi
	echo $GOOS $GOARCH $EXT
	go build -o mog-$GOOS-$GOARCH$EXT
}

build windows amd64
build windows 386
build linux amd64
