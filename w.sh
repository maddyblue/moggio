#!/bin/sh

set -e

while true; do
	go run main.go -w -dev
	if [ $? != 0 ] ; then
		exit
	fi
	echo restarting
done