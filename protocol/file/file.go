package file

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/protocol"
)

func init() {
	protocol.Register("file", []string{"directory"}, List)
}

func List(params []string) (protocol.SongList, error) {
	if len(params) != 1 {
		return nil, fmt.Errorf("bad params")
	}

	f, err := os.Open(params[0])
	if err != nil {
		return nil, err
	}
	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", params[0])
	}
	songs := make(protocol.SongList)
	var walk func(string)
	walk = func(dirname string) {
		f, err := os.Open(dirname)
		if err != nil {
			return
		}
		fis, err := f.Readdir(0)
		if err != nil {
			return
		}
		for _, fi := range fis {
			p := filepath.Join(dirname, fi.Name())
			if fi.IsDir() {
				walk(p)
			} else {
				f, err := os.Open(p)
				if err != nil {
					continue
				}
				ss, _, err := codec.Decode(f)
				if err != nil {
					continue
				}
				for i, s := range ss {
					id := fmt.Sprintf("%v-%v", i, p)
					songs[id] = s
				}
			}
		}
	}
	walk(params[0])
	return songs, nil
}
