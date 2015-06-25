vorbis
======

[![GoDoc](https://godoc.org/github.com/mccoyst/vorbis?status.svg)](https://godoc.org/github.com/mccoyst/vorbis)
[![Build Status](https://travis-ci.org/mccoyst/vorbis.svg?branch=master)](https://travis-ci.org/mccoyst/vorbis)

This Go package provides a "native" ogg vorbis decoder, but still requires cgo, as it uses inline code from [stb_vorbis](http://nothings.org/stb_vorbis/). Someday, it won't.

The package exports a single function:

	var data []byte
	…
	samples, nchannels, sampleRate, err := vorbis.Decode(data)
	
This corresponds to `stb_vorbis_decode_memory()`, but is a little different. Samples is a `[]int16`, corresponding to stb's dynamic array of shorts if you're on the right platforms. The samples seem to be stored native-endian, but I haven't tested many vorbis files. Nchannels is the number of channels, which are interleaved in the samples slice. Err is non-nil if the data is not an ogg vorbis stream according to stb.
