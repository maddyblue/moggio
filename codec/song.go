package codec

import (
	"io"
	"time"
)

type Song interface {
	io.Reader
	// Info returns information about a song.
	Info() (SongInfo, error)
	// Init is called before the first call to Play(). It should prepare resources
	// needed for Play().
	Init() (sampleRate, channels int, err error)
	// Close releases resources used by the current file.
	Close()
}

type SongInfo struct {
	Time     time.Duration
	Artist   string
	Title    string
	Album    string
	Track    float64
	ImageURL string `json:",omitempty"`

	// SongTitle, if set, is the currently playing song title. Needed for
	// streaming.
	SongTitle string
}
