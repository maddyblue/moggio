// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package codec

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dhowden/tag"
)

// ErrFormat indicates that decoding encountered an unknown format.
var ErrFormat = errors.New("codec: unknown format")

type Songs map[ID]Song

type ID string

const None ID = ""

const IdSep = "\n"

func NewID(s ...string) ID {
	return ID(strings.Join(s, IdSep))
}

func Int(i int) ID {
	return ID(strconv.Itoa(i))
}

func Int64(i int64) ID {
	return ID(strconv.FormatInt(i, 10))
}

func (i ID) Top() string {
	return strings.SplitN(string(i), IdSep, 2)[0]
}

func (i ID) Pop() (string, ID) {
	s := strings.SplitN(string(i), IdSep, 2)
	if len(s) == 1 {
		return s[0], ""
	}
	return s[0], ID(s[1])
}

func (i ID) Push(s string) ID {
	return i + IdSep + ID(s)
}

type codec struct {
	name       string
	magic      []string
	extensions []string
	decode     func(Reader) (Songs, error)
	get        func(Reader, ID) (Song, error)
}

var (
	// Codecs is the list of registered codecs.
	codecs        = make(map[string]*codec)
	allExtensions = make(map[string]*codec)
)

// RegisterCodec registers an audio codec for use by Decode.
// Name is the name of the format, like "nsf" or "wav".
// Magic is the magic prefix that identifies the codec's encoding. The magic
// string can contain "?" wildcards that each match any one byte.
// Decode is the function that decodes the encoded codec.
func RegisterCodec(name string, magic, extensions []string, decode func(Reader) (Songs, error), get func(Reader, ID) (Song, error)) {
	if _, ok := codecs[name]; ok {
		panic(fmt.Errorf("%v already registered", name))
	}
	c := &codec{
		name:       name,
		magic:      magic,
		extensions: extensions,
		decode:     decode,
		get:        get,
	}
	for _, e := range extensions {
		if v, ok := allExtensions[e]; ok {
			panic(fmt.Errorf("%v already owns extension %v", v.name, e))
		}
		allExtensions[e] = c
	}
	codecs[name] = c
}

// A reader is an io.Reader that can also peek ahead.
type reader interface {
	io.Reader
	Peek(int) ([]byte, error)
}

// asReader converts an io.Reader to a reader.
func asReader(r io.Reader) reader {
	if rr, ok := r.(reader); ok {
		return rr
	}
	return bufio.NewReader(r)
}

// Match returns whether magic matches b. Magic may contain "?" wildcards.
func match(magic string, b []byte) bool {
	if len(magic) != len(b) {
		return false
	}
	for i, c := range b {
		if magic[i] != c && magic[i] != '?' {
			return false
		}
	}
	return true
}

// Sniff determines the format of r's data.
func sniff(r reader) *codec {
	for _, c := range codecs {
		for _, m := range c.magic {
			b, err := r.Peek(len(m))
			if err == nil && match(m, b) {
				return c
			}
		}
	}
	return nil
}

// Reader returns a file reader and the file size in bytes (or 0 if streamed
// or unknown).
type Reader func() (io.ReadCloser, int64, error)

func (rf Reader) Metadata(ft tag.FileType) (*SongInfo, tag.Metadata, []byte, error) {
	r, sz, err := rf()
	if err != nil {
		return nil, nil, nil, err
	}
	if sz == 0 {
		return nil, nil, nil, fmt.Errorf("cannot get metadata with unknown size")
	}
	b, err := ioutil.ReadAll(r)
	r.Close()
	if err != nil {
		return nil, nil, nil, err
	}
	m, err := tag.ReadFrom(bytes.NewReader(b))
	if err != nil {
		return nil, nil, nil, err
	}
	if m.FileType() != ft {
		return nil, nil, nil, fmt.Errorf("expected filetype %v, got %v", ft, m.FileType())
	}
	track, _ := m.Track()
	si := &SongInfo{
		Artist:   m.Artist(),
		Title:    m.Title(),
		Album:    m.Album(),
		Track:    float64(track),
		ImageURL: dataURL(m),
	}
	return si, m, b, nil
}

func dataURL(m tag.Metadata) string {
	p := m.Picture()
	if p == nil {
		return ""
	}
	if p.MIMEType == "-->" {
		return string(p.Data)
	}
	return fmt.Sprintf("data:%s;base64,%s", p.MIMEType, base64.StdEncoding.EncodeToString(p.Data))
}

// Decode decodes audio that has been encoded in a registered codec.
// The string returned is the format name used during format registration.
// Format registration is typically done by the init method of the codec-
// specific package.
func Decode(rf Reader) (Songs, string, error) {
	r, _, err := rf()
	if err != nil {
		return nil, "", err
	}
	defer r.Close()
	rr := asReader(r)
	f := sniff(rr)
	if f == nil {
		return nil, "", ErrFormat
	}
	m, err := f.decode(rf)
	return m, f.name, err
}

func extension(path string) (*codec, error) {
	ext := filepath.Ext(path)
	ext = strings.Trim(ext, ".")
	if ext == "" {
		ext = path
	}
	c, ok := allExtensions[ext]
	if !ok {
		return nil, fmt.Errorf("extension not found: %v", ext)
	}
	return c, nil
}

func ByExtension(path string, rf Reader) (Songs, string, error) {
	c, err := extension(path)
	if err != nil {
		return nil, "", err
	}
	songs, err := c.decode(rf)
	return songs, c.name, err
}

func ByExtensionID(path string, id ID, rf Reader) (Song, error) {
	c, err := extension(path)
	if err != nil {
		return nil, err
	}
	if c.get != nil {
		return c.get(rf, id)
	}
	songs, err := c.decode(rf)
	if err != nil {
		return nil, err
	}
	song, ok := songs[id]
	if !ok {
		return nil, fmt.Errorf("song not found: %v", id)
	}
	return song, nil
}
