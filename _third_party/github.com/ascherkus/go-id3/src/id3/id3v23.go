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
	"bufio"
	"encoding/binary"
)

// ID3 v2.3 doesn't use sync-safe frame sizes: read in as a regular big endian number.
func parseID3v23Size(reader *bufio.Reader) int {
	var size int32
	binary.Read(reader, binary.BigEndian, &size)
	return int(size)
}

func parseID3v23File(reader *bufio.Reader, file *File) {
	for hasFrame(reader, 4) {
		id := string(readBytes(reader, 4))
		size := parseID3v23Size(reader)

		// Skip over frame flags.
		skipBytes(reader, 2)

		switch id {
		case "TALB":
			file.Album = readString(reader, size)
		case "TRCK":
			file.Track = readString(reader, size)
		case "TPE1":
			file.Artist = readString(reader, size)
		case "TCON":
			file.Genre = readGenre(reader, size)
		case "TIT2":
			file.Name = readString(reader, size)
		case "TYER":
			file.Year = readString(reader, size)
		case "TPOS":
			file.Disc = readString(reader, size)
		case "TLEN":
			file.Length = readString(reader, size)
		default:
			skipBytes(reader, size)
		}
	}
}
