// Copyright 2011 Andrew Scherkus
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package id3

import (
	"bytes"
	"os"
	"path"
	"testing"
)

type fileTest struct {
	path string
	file File
}

func testFile(t *testing.T, expected fileTest) {
	p := path.Join("..", "..", "test", expected.path)
	t.Logf("With file %s", p)

	fd, err := os.Open(p)
	if fd == nil || err != nil {
		t.Error(err)
		return
	}
	defer fd.Close()

	actual := Read(fd)
	if actual == nil {
		t.Error("Could not parse ID3 information")
		return
	}

	// ID3v2Header fields.
	header := expected.file.Header
	if actual.Header.Version != header.Version {
		t.Errorf("Header.Version: expected %d got %d", header.Version, actual.Header.Version)
	}
	if actual.Header.MinorVersion != header.MinorVersion {
		t.Errorf("Header.MinorVersion: expected %d got %d", header.MinorVersion, actual.Header.MinorVersion)
	}
	if actual.Header.Unsynchronization != header.Unsynchronization {
		t.Errorf("Header.Unsynchronization: expected %t got %t", header.Unsynchronization, actual.Header.Unsynchronization)
	}
	if actual.Header.Extended != header.Extended {
		t.Errorf("Header.Extended: expected %t got %t", header.Extended, actual.Header.Extended)
	}
	if actual.Header.Experimental != header.Experimental {
		t.Errorf("Header.Experimental: expected %t got %t", header.Experimental, actual.Header.Experimental)
	}
	if actual.Header.Footer != header.Footer {
		t.Errorf("Header.Footer: expected %t got %t", header.Footer, actual.Header.Footer)
	}
	if actual.Header.Size != header.Size {
		t.Errorf("Header.Size: expected %d got %d", header.Size, actual.Header.Size)
	}

	// Name, Artist, etc...
	file := expected.file
	if actual.Name != file.Name {
		t.Errorf("Name: expected '%s' got '%s'", file.Name, actual.Name)
	}
	if actual.Artist != file.Artist {
		t.Errorf("Artist: expected '%s' got '%s'", file.Artist, actual.Artist)
	}
	if actual.Album != file.Album {
		t.Errorf("Album: expected '%s' got '%s'", file.Album, actual.Album)
	}
	if actual.Year != file.Year {
		t.Errorf("Year: expected '%s' got '%s'", file.Year, actual.Year)
	}
	if actual.Track != file.Track {
		t.Errorf("Track: expected '%s' got '%s'", file.Track, actual.Track)
	}
	if actual.Disc != file.Disc {
		t.Errorf("Disc: expected '%s' got '%s'", file.Disc, actual.Disc)
	}
	if actual.Genre != file.Genre {
		t.Errorf("Genre: expected '%s' got '%s'", file.Genre, actual.Genre)
	}
	if actual.Length != file.Length {
		t.Errorf("Length: expected '%s' got '%s'", file.Length, actual.Length)
	}
}

func TestEmpty(t *testing.T) {
	file := Read(new(bytes.Buffer))
	if file != nil {
		t.Fail()
	}
}

func TestID3v220(t *testing.T) {
	testFile(t, fileTest{"test_220.mp3", File{ID3v2Header{2, 0, false, false, false, false, 226741},
		"There There", "Radiohead", "Hail To The Thief", "2003", "9", "", "Alternative", ""}})
}

func TestID3v230(t *testing.T) {
	testFile(t, fileTest{"test_230.mp3", File{ID3v2Header{3, 0, false, false, false, false, 150717},
		"Everything In Its Right Place", "Radiohead", "Kid A", "2000", "1", "", "Alternative", ""}})
}

func TestID3v240(t *testing.T) {
	testFile(t, fileTest{"test_240.mp3", File{ID3v2Header{4, 0, false, false, false, false, 165126},
		"Give Up The Ghost", "Radiohead", "The King Of Limbs", "2011", "07/08", "1/1", "Alternative", ""}})
}

func TestISO8859_1(t *testing.T) {
	testFile(t, fileTest{"test_iso8859_1.mp3", File{ID3v2Header{3, 0, false, false, false, false, 273649},
		"Pompeii Am Götterdämmerung", "The Flaming Lips", "At War With The Mystics", "2006", "11", "1/1", "Unknown", ""}})
}
