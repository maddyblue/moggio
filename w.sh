#!/bin/sh

set -e

while true; do
	time go install
	GOTRACEBACK=all moggio -w -dev
	if [ $? != 0 ] ; then
		exit
	fi
	echo restarting
done