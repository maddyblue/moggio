// mpa, an MPEG-1 Audio library
// Copyright (C) 2014 KORÁNDI Zoltán <korandi.z@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License, version 3 as
// published by the Free Software Foundation.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// Please note that, being hungarian, my last name comes before my first
// name. That's why it's in all caps, and not because I like to shout my
// name. So please don't start your emails with "Hi Korandi" or "Dear Mr.
// Zoltan", because it annoys the hell out of me. Thanks.

// Package mpa is an MPEG-1 Audio library. It's currently decoding-only.
//
// A trivial example which reads the MPEG-1 coded bitstream from stdin and
// writes the decoded PCM stream to stdout:
//
//   package main
//
//   import (
//       "io"
//       "os"
//
//       "github.com/korandiz/mpa"
//   )
//
//   func main() {
//       io.Copy(os.Stdout, &mpa.Reader{Decoder: &mpa.Decoder{Input: os.Stdin}})
//   }
//
// On Linux, for example, if you have alsa-utils installed, you can play an mp3
// using the above code with something like this:
//
//     go run decode.go < nyan.mp3 | aplay -fcd
//
package mpa
