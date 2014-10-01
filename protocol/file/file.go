package file

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/protocol"
)

func init() {
	protocol.Register("file", []string{"directory"}, List)
}

func List(inst *protocol.Instance) (protocol.SongList, error) {
	if len(inst.Params) != 1 {
		return nil, fmt.Errorf("file: bad params")
	}
	songs := make(protocol.SongList)
	err := filepath.Walk(inst.Params[0], func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		f, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer f.Close()
		ss, _, err := codec.Decode(fileReader(path))
		if err != nil {
			return nil
		}
		for i, s := range ss {
			id := fmt.Sprintf("%v-%v", i, path)
			songs[id] = s
		}
		return nil
	})
	return songs, err
}

func fileReader(path string) codec.Reader {
	return func() (io.ReadCloser, int64, error) {
		log.Println("open file", path)
		f, err := os.Open(path)
		if err != nil {
			return nil, 0, err
		}
		fi, err := f.Stat()
		if err != nil {
			f.Close()
			return nil, 0, err
		}
		return f, fi.Size(), nil
	}
}
