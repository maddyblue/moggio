#!/bin/sh

set -e

while true; do
	time go install
	GOTRACEBACK=all mog -w -dev
	if [ $? != 0 ] ; then
		exit
	fi
	echo restarting
done