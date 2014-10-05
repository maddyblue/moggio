#!/bin/sh

while true; do
	go run main.go -w -dropbox rnhpqsbed2q2ezn:ldref688unj74ld -drive 256229448371-93bchgphf79q2vbik5aod4osvksce35p.apps.googleusercontent.com:zO5-2BqMb5Zl4EFKd3fVnavw -soundcloud ec28c2226a0838d01edc6ed0014e462e:a115e94029d698f541960c8dc8560978
	if [ $? != 0 ] ; then
		exit
	fi
	echo restarting
done