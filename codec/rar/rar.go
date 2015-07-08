package rar

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/mjibson/mog/_third_party/github.com/nwaples/rardecode"
	"github.com/mjibson/mog/codec"
)

func init() {
	codec.RegisterCodec("RAR", []string{"RAR!\u001a\u0007"}, []string{"rar", "rsn"}, Read, Get)
}

func read(r io.Reader, f func(*rardecode.Reader, *rardecode.FileHeader) (stop bool)) error {
	d, err := rardecode.NewReader(r, "")
	if err != nil {
		return err
	}
	for {
		fh, err := d.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		} else if fh.IsDir {
			continue
		}
		if f(d, fh) {
			break
		}
	}
	return nil
}

func Read(rf codec.Reader) (codec.Songs, error) {
	r, _, err := rf()
	if err != nil {
		return nil, err
	}
	// In order to not open the rar file up once per included file, keep around a
	// pointer to the reader during the inital listing phase and use that if we're
	// opening it. I would appreciate a better way to do this.
	var rd *rardecode.Reader
	var rfh *rardecode.FileHeader
	defer func() {
		rd = nil
	}()
	readRAR := func(name string, rf codec.Reader) codec.Reader {
		return func() (rc io.ReadCloser, sz int64, err error) {
			if rd != nil && rfh.Name == name {
				return ioutil.NopCloser(rd), rfh.UnPackedSize, nil
			}
			r, _, err := rf()
			if err != nil {
				return nil, 0, err
			}
			f := func(d *rardecode.Reader, fh *rardecode.FileHeader) (stop bool) {
				if fh.Name != name {
					return false
				}
				rc = &rarReader{r, d}
				sz = fh.UnPackedSize
				return true
			}
			err = read(r, f)
			if err == nil && rc == nil {
				err = fmt.Errorf("rar: %v unfound", name)
			}
			return
		}
	}
	defer r.Close()
	songs := make(codec.Songs)
	f := func(d *rardecode.Reader, fh *rardecode.FileHeader) (stop bool) {
		rd = d
		rfh = fh
		ss, _, _ := codec.ByExtension(fh.Name, readRAR(fh.Name, rf))
		for v, s := range ss {
			songs[codec.NewID(fh.Name, string(v))] = s
		}
		return false
	}
	if err := read(r, f); err != nil {
		return nil, err
	}
	return songs, nil
}

type id struct {
	name string
	v    interface{}
}

type rarReader struct {
	file   io.ReadCloser
	reader io.Reader
}

func (r *rarReader) Close() error {
	return r.file.Close()
}

func (r *rarReader) Read(p []byte) (int, error) {
	return r.reader.Read(p)
}

func Get(rf codec.Reader, id codec.ID) (codec.Song, error) {
	top, child := id.Pop()
	r, _, err := rf()
	if err != nil {
		return nil, err
	}
	defer r.Close()
	var song codec.Song
	f := func(d *rardecode.Reader, fh *rardecode.FileHeader) (stop bool) {
		if fh.Name != top {
			return false
		}
		var b []byte
		// TODO: Wait until needed to read the data. However, this should cache the data
		// and then close the reader.
		b, err = ioutil.ReadAll(d)
		if err != nil {
			return true
		}
		song, err = codec.ByExtensionID(top, child, func() (io.ReadCloser, int64, error) {
			return ioutil.NopCloser(bytes.NewReader([]byte(b))), int64(len(b)), nil
		})
		return true
	}
	err = read(r, f)
	if err == nil && song == nil {
		err = fmt.Errorf("rar: %v unfound", top)
	}
	return song, err
}
