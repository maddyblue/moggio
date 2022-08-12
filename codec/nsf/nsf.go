//go:build !cgo
// +build !cgo

package nsf

import (
	"fmt"
	"strconv"

	"github.com/mjibson/moggio/codec"
	"github.com/mjibson/nsf"
)

func init() {
	codec.RegisterCodec("NSF", []string{"NESM\u001a"}, []string{"nsf"}, ReadNSFSongs, Get)
	codec.RegisterCodec("NSFE", []string{"NSFE"}, []string{"nsfe"}, ReadNSFSongs, Get)
}

func ReadNSFSongs(rf codec.Reader) (codec.Songs, error) {
	r, _, err := rf()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	n, err := nsf.New(r)
	if err != nil {
		return nil, err
	}
	songs := make(codec.Songs, len(n.Songs))
	for i := 0; i < len(n.Songs); i++ {
		songs[codec.Int(i)] = &NSFSong{
			NSF:    n,
			Index:  i + 1,
			Reader: rf,
		}
	}
	return songs, nil
}

func Get(rf codec.Reader, id codec.ID) (codec.Song, error) {
	i, err := strconv.Atoi(string(id))
	if err != nil {
		return nil, err
	}
	return &NSFSong{
		Index:  i + 1,
		Reader: rf,
	}, nil
}

type NSFSong struct {
	NSF     *nsf.NSF
	Index   int
	Playing bool
	Reader  codec.Reader
}

func (n *NSFSong) Init() (sampleRate, channels int, err error) {
	if n.NSF == nil {
		r, _, err := n.Reader()
		if err != nil {
			return 0, 0, err
		}
		defer r.Close()
		n.NSF, err = nsf.New(r)
		if err != nil {
			return 0, 0, err
		}
	}
	n.NSF.Init(n.Index)
	n.Playing = true
	return int(n.NSF.SampleRate), 1, nil
}

func (n *NSFSong) Play(samples int) ([]float32, error) {
	return n.NSF.Play(samples), nil
}

func (n *NSFSong) Close() {
	n.NSF = nil
	n.Playing = false
}

func (n *NSFSong) Info() (si codec.SongInfo, err error) {
	ns := n.NSF
	if ns == nil {
		r, _, err := n.Reader()
		if err != nil {
			return si, err
		}
		defer r.Close()
		ns, err = nsf.New(r)
		if err != nil {
			return si, err
		}
	}
	s := n.NSF.Songs[n.Index-1]
	title := s.Name
	if title == "" {
		title = fmt.Sprintf("%s:%02d", n.NSF.Game, n.Index)
	}
	si = codec.SongInfo{
		Time:   s.Duration,
		Artist: n.NSF.Artist,
		Album:  n.NSF.Game,
		Track:  float64(n.Index),
		Title:  title,
	}
	return
}
