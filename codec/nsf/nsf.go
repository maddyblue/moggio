package nsf

import (
	"fmt"
	"io"
	"time"

	"github.com/mjibson/mog/codec"
	"github.com/mjibson/nsf"
)

func init() {
	codec.RegisterCodec("NSF", "NESM\u001a", ReadNSFSongs)
}

func ReadNSFSongs(r io.Reader) ([]codec.Song, error) {
	n, err := nsf.ReadNSF(r)
	if err != nil {
		return nil, err
	}
	songs := make([]codec.Song, n.Songs)
	for i := range songs {
		songs[i] = &NSFSong{
			NSF:   n,
			Index: i + 1,
		}
	}
	return songs, nil
}

type NSFSong struct {
	*nsf.NSF
	Index   int
	Playing bool
}

func (n *NSFSong) Play(samples int) []float32 {
	if !n.Playing {
		n.Init(n.Index)
		n.Playing = true
	}
	return n.NSF.Play(samples)
}

func (n *NSFSong) Close() {
	// todo: implement
}

func (n *NSFSong) Info() codec.SongInfo {
	return codec.SongInfo{
		Time:       time.Minute * 2,
		Artist:     n.Artist,
		Album:      n.Song,
		Track:      n.Index,
		Title:      fmt.Sprintf("%s:%d", n.Song, n.Index),
		SampleRate: int(n.SampleRate),
		Channels:   1,
	}
}
