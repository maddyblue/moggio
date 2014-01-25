# mog

An MPD-compatible audio player written in Go.

# goals

1. Useful compatibility with the MPD protocol. Perhaps the full protocol isn't supported, but most MPD clients will work with mog.
1. Support for the following formats:
   * wav
   * ogg
   * flac
   * mp3
   * spc (Super Nintendo)
   * nsf, nsfe (Nintendo)
1. Support for the Google Music protocol.
1. Support for archive files (.zip, .rar, .nsf, etc.).
1. Pure Go except for sound driver interfaces.
1. Runs on Windows, Linux, Mac OSX.
