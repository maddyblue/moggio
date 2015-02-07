#!/bin/sh

set -e

docker build -t mjibson/mog .

DIR=/go/src/github.com/mjibson/mog
docker run --rm -v "$(pwd)":$DIR -w $DIR mjibson/mog sh docker.sh