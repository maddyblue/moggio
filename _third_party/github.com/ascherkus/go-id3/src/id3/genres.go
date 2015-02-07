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
	"fmt"
	"strconv"
	"strings"
)

var id3v1Genres = []string{
	"Blues",
	"Classic Rock",
	"Country",
	"Dance",
	"Disco",
	"Funk",
	"Grunge",
	"Hip-Hop",
	"Jazz",
	"Metal",
	"New Age",
	"Oldies",
	"Other",
	"Pop",
	"R&B",
	"Rap",
	"Reggae",
	"Rock",
	"Techno",
	"Industrial",
	"Alternative",
	"Ska",
	"Death Metal",
	"Pranks",
	"Soundtrack",
	"Euro-Techno",
	"Ambient",
	"Trip-Hop",
	"Vocal",
	"Jazz+Funk",
	"Fusion",
	"Trance",
	"Classical",
	"Instrumental",
	"Acid",
	"House",
	"Game",
	"Sound Clip",
	"Gospel",
	"Noise",
	"AlternRock",
	"Bass",
	"Soul",
	"Punk",
	"Space",
	"Meditative",
	"Instrumental Pop",
	"Instrumental Rock",
	"Ethnic",
	"Gothic",
	"Darkwave",
	"Techno-Industrial",
	"Electronic",
	"Pop-Folk",
	"Eurodance",
	"Dream",
	"Southern Rock",
	"Comedy",
	"Cult",
	"Gangsta",
	"Top 40",
	"Christian Rap",
	"Pop/Funk",
	"Jungle",
	"Native American",
	"Cabaret",
	"New Wave",
	"Psychadelic",
	"Rave",
	"Showtunes",
	"Trailer",
	"Lo-Fi",
	"Tribal",
	"Acid Punk",
	"Acid Jazz",
	"Polka",
	"Retro",
	"Musical",
	"Rock & Roll",
	"Hard Rock",
}

// ID3v2.2 and ID3v2.3 use "(NN)" where as ID3v2.4 simply uses "NN" when
// referring to ID3v1 genres. The "(NN)" format is allowed to have trailing
// information.
//
// RX and CR are shorthand for Remix and Cover, respectively.
//
// Refer to the following documentation:
//   http://id3.org/id3v2-00          TCO frame
//   http://id3.org/id3v2.3.0         TCON frame
//   http://id3.org/id3v2.4.0-frames  TCON frame
func convertID3v1Genre(genre string) string {
	if genre == "RX" || strings.HasPrefix(genre, "(RX)") {
		return "Remix"
	}
	if genre == "CR" || strings.HasPrefix(genre, "(CR)") {
		return "Cover"
	}

	// Try to parse "NN" format.
	index, err := strconv.Atoi(genre)
	if err == nil {
		if index >= 0 && index < len(id3v1Genres) {
			return id3v1Genres[index]
		}
		return "Unknown"
	}

	// Try to parse "(NN)" format.
	index = 0
	_, err = fmt.Sscanf(genre, "(%d)", &index)
	if err == nil {
		if index >= 0 && index < len(id3v1Genres) {
			return id3v1Genres[index]
		}
		return "Unknown"
	}

	// Couldn't parse so it's likely not an ID3v1 genre.
	return genre
}
