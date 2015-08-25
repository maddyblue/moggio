#!/bin/sh

set -e

while true; do
	time go install
	mog -w -dev
	if [ $? != 0 ] ; then
		exit
	fi
	echo restarting
done