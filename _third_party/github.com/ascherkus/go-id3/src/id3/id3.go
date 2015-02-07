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

// Package id3 implements basic ID3 parsing for MP3 files.
//
// Instead of providing access to every single ID3 frame this package
// exposes only the ID3v2 header and a few basic fields such as the
// artist, album, year, etc...
package id3

import (
	"bufio"
	"fmt"
	"io"
)

// A parsed ID3v2 header as defined in Section 3 of
// http://id3.org/id3v2.4.0-structure
type ID3v2Header struct {
	Version           int
	MinorVersion      int
	Unsynchronization bool
	Extended          bool
	Experimental      bool
	Footer            bool
	Size              int32
}

// A parsed ID3 file with common fields exposed.
type File struct {
	Header ID3v2Header

	Name   string
	Artist string
	Album  string
	Year   string
	Track  string
	Disc   string
	Genre  string
	Length string
}

// Parse the input for ID3 information. Returns nil if parsing failed or the
// input didn't contain ID3 information.
func Read(reader io.Reader) *File {
	file := new(File)
	bufReader := bufio.NewReader(reader)
	if !isID3Tag(bufReader) {
		return nil
	}

	parseID3v2Header(bufReader, file)
	limitReader := bufio.NewReader(io.LimitReader(bufReader, int64(file.Header.Size)))
	if file.Header.Version == 2 {
		parseID3v22File(limitReader, file)
	} else if file.Header.Version == 3 {
		parseID3v23File(limitReader, file)
	} else if file.Header.Version == 4 {
		parseID3v24File(limitReader, file)
	} else {
		panic(fmt.Sprintf("Unrecognized ID3v2 version: %d", file.Header.Version))
	}

	return file
}

func isID3Tag(reader *bufio.Reader) bool {
	data, err := reader.Peek(3)
	if len(data) < 3 || err != nil {
		return false
	}
	return data[0] == 'I' && data[1] == 'D' && data[2] == '3'
}

func parseID3v2Header(reader *bufio.Reader, file *File) {
	data := readBytes(reader, 10)
	file.Header.Version = int(data[3])
	file.Header.MinorVersion = int(data[4])
	file.Header.Unsynchronization = data[5]&1<<7 != 0
	file.Header.Extended = data[5]&1<<6 != 0
	file.Header.Experimental = data[5]&1<<5 != 0
	file.Header.Footer = data[5]&1<<4 != 0
	file.Header.Size = parseSize(data[6:])
}
