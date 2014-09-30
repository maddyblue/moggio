package nsf

import (
	"fmt"
	"time"

	"github.com/mjibson/mog/codec"
	"github.com/mjibson/nsf"
)

func init() {
	codec.RegisterCodec("NSF", "NESM\u001a", ReadNSFSongs)
}

func ReadNSFSongs(rf codec.Reader) ([]codec.Song, error) {
	r, err := rf()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	n, err := nsf.ReadNSF(r)
	if err != nil {
		return nil, err
	}
	songs := make([]codec.Song, n.Songs)
	for i := range songs {
		songs[i] = &NSFSong{
			Index:  i + 1,
			Reader: rf,
		}
	}
	return songs, nil
}

type NSFSong struct {
	NSF     *nsf.NSF
	Index   int
	Playing bool
	Reader  codec.Reader
}

func (n *NSFSong) Init() (sampleRate, channels int, err error) {
	r, err := n.Reader()
	if err != nil {
		return 0, 0, err
	}
	defer r.Close()
	n.NSF, err = nsf.ReadNSF(r)
	if err != nil {
		return 0, 0, err
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
	if n.NSF != nil {
		si = codec.SongInfo{
			Time:   time.Minute * 2,
			Artist: n.NSF.Artist,
			Album:  n.NSF.Song,
			Track:  float64(n.Index),
			Title:  fmt.Sprintf("%s:%d", n.NSF.Song, n.Index),
		}
		return
	}
	r, err := n.Reader()
	if err != nil {
		return
	}
	defer r.Close()
	ns, err := nsf.ReadNSF(r)
	if err != nil {
		return
	}
	si = codec.SongInfo{
		Time:   time.Minute * 2,
		Artist: ns.Artist,
		Album:  ns.Song,
		Track:  float64(n.Index),
		Title:  fmt.Sprintf("%s:%d", ns.Song, n.Index),
	}
	return
}
