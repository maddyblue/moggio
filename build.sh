#!/bin/sh

set -e

if [ -z "$1" ]; then
	echo "Must specify version"
	exit
fi

docker build -t mjibson/mog .

DIR=/go/src/github.com/mjibson/mog
docker run --rm -v "$(pwd)":$DIR -w $DIR mjibson/mog sh docker.sh $1
