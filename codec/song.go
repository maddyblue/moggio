package codec

import "time"

type Song interface {
	// Info returns information about a song.
	Info() (SongInfo, error)
	// Init is called before the first call to Play(). It should prepare resources
	// needed for Play().
	Init() (sampleRate, channels int, err error)
	// Play returns the next n samples. Return < n to indicate end of song or 0
	// to indicate EOF, neither of which should return an error. If < n samples
	// are returned, Play will not be invoked again.
	Play(n int) ([]float32, error)
	// Close releases resources used by the current file.
	Close()
}

type SongInfo struct {
	Time      time.Duration
	Artist    string
	Title     string
	Album     string
	Track     float64
	ImageURL  string `json:",omitempty"`
}
