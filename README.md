# mog

A modern audio player written in Go.

# goals

1. JSON API. MPD may be supported at some later date, but is not a goal.
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
