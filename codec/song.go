package codec

import "time"

type Song interface {
	// Info returns information about a song.
	Info() SongInfo
	// Play returns the next n samples.
	Play(n int) []float32
	// Close releases resources used by the current file. The next call to Play()
	// will reopen the song at 0:00.
	Close()
}

type SongInfo struct {
	Time   time.Duration
	Artist string
	Title  string
	Album  string
	Track  int
	SampleRate int
}
