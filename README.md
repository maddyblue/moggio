# mog

An audio player written in Go.

# goals

1. JSON API. Playlists easily managed with a web browser, and global OS keyboard shortcuts can access functions needed from the media keys (next, pause, play).
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

# current status

- When run, can play your MP3 and NSF files from Dropbox, Google Drive, Soundcloud, local hard drive, and Google Music.
- UI is improving but usable enough.
- Connect to [http://localhost:6601](http://localhost:6601) to see the UI.
- Dropbox, Google Drive, and Soundcloud use my API keys by default, but can be changed on the command line. Their oauth redirects go to `localhost:6601`.
