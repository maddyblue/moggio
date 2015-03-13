# mog

An audio player written in Go.

# goals

1. JSON API. Playlists easily managed with a web browser, and global OS keyboard shortcuts can access functions needed from the media keys (next, pause, play).
1. Support for the codecs:
   * wav
   * mp3
   * spc (Super Nintendo)
   * nsf, nsfe (Nintendo)
   * aac
1. Support for the protocols:
   * google music
   * dropbox
   * google drive
   * shoutcast
   * soundcloud
   * local hard drive
1. Support for archive files (.zip, .rar, .nsf).
1. Pure Go except for sound driver interfaces.
1. Runs on Windows, Linux, Mac OSX.
1. Mobile apps for Android and iOS that can play themselves are act as a remote.
