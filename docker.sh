#!/bin/sh

set -e

ver=$1

build()
{
	export GOOS=$1
	export GOARCH=$2
	EXT=""
	if [ $GOOS = "windows" ]; then
		EXT=".exe"
	fi
	echo $GOOS $GOARCH $EXT
	go build -o moggio-$GOOS-$GOARCH$EXT -ldflags "-X github.com/mjibson/moggio/server.MoggioVersion=$ver"
}

go version
build windows amd64
build linux amd64
