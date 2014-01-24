/*
Package nsf provides reading and emulating of Nintendo NSF sound files.

PortAudio is the current default for audio output.

To install PortAudio on Mac OSX:

	brew install portaudio

To install PortAudio on Ubuntu:

	sudo apt-get install portaudio19-dev

Then play with go test. The track can be changed by incrementing the parameter
in the n.Init() call in nsf_test.go.
*/
package nsf
