#!/bin/sh

set -e

if [ -z "$1" ]; then
	echo "Must specify version"
	exit
fi

docker build -t mjibson/moggio .

DIR=/go/src/github.com/mjibson/moggio
docker run --rm -v "$(pwd)":$DIR -w $DIR mjibson/moggio sh docker.sh $1
