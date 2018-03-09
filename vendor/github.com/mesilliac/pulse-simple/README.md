pulse-simple
============

Cgo bindings to PulseAudio's Simple API,
for easily playing or capturing raw audio.

The full Simple API is supported,
including channel mapping and setting buffer attributes.

prerequisites
-------------

These bindings require the pulseaudio C headers to be available.
On Ubuntu they can be installed by `sudo apt install libpulse-dev`,
and on other distros there should be a similar package available.

quick test
----------

If everything is configured correctly,
`go run examples/sinewave.go` should output some simple tones via pulseaudio.

usage
-----

Basic usage is to request a playback or capture stream,
then write bytes to or read bytes from it.

Reading and writing will block until the given byte slice
is completely consumed or filled, or an error occurs.

The format of the data will be as requested on stream creation.

For example,
assuming "data" contains raw bytes representing stereophonic audio
in little-endian 16-bit integer PCM format,
the following will obtain a playback stream
and play the given data as audio on the default sound device.

    ss := pulse.SampleSpec{pulse.SAMPLE_S16LE, 44100, 2}
    stream, _ := pulse.Playback("my app", "my stream", &ss)
    defer stream.Free()
    defer stream.Drain()
    stream.Write(data)

More example usage can be found in the examples folder.

For more information, see the PulseAudio Simple API documentation at
http://www.freedesktop.org/software/pulseaudio/doxygen/simple.html

license
-------

MIT (see the included LICENSE file for full license text)

