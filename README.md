# moggio

An audio player written in Go.

# goals

- [x] JSON API. Playlists easily managed with a web browser, and global OS keyboard shortcuts can access functions needed from the media keys (next, pause, play).
- [ ] Support for the codecs:
  - [x] wav
  - [x] mp3
  - [x] spc (Super Nintendo)
  - [x] nsf, nsfe (Nintendo)
  - [x] ogg vorbis
  - [x] flac
  - [ ] aac
- [ ] Support for the protocols:
  - [x] google music
  - [x] dropbox
  - [x] google drive
  - [x] shoutcast
  - [x] soundcloud
  - [x] local hard drive
  - [ ] youtube
- [x] Support for archive files (.zip, .rar, .nsf).
- [ ] Pure Go except for sound driver interfaces.
  - [ ] Port SPC library to Go
- [x] Runs on Windows, Linux, Mac OSX.
- [ ] Mobile apps for Android and iOS that can play themselves and act as a remote.
