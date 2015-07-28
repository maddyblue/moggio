// mpseek, a library to support seeking MPEG Audio files
// Copyright (C) 2015 KORÁNDI Zoltán <korandi.z@gmail.com>
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

// Package mpseek provides seek tables for MPEG-1 Audio files.
//
// Seeking in an mp3 is tricky. The stream consists of so-called frames, which
// all code the same number of samples (and the same length of sound), but their
// size in bytes can vary, even when the bitrate is nominally constant. This
// means you can't use some simple formula to map seconds to byte offsets.
// Moreover, frames carry no timestamps, so you can't use bisection search
// either.
//
// The only viable solution is to parse the stream and precompute a list of
// (timestamp, offset) pairs. Such a list is called a seek table. There is an
// obvious tradeoff here: The more seek points you add, the more accurate your
// seeks will be, but a larger table will take up more memory.
//
// To make life more complicated, the decoder needs to be "warmed up" before it
// can provide correct output, so decoding must start a few frames before the
// seek point. The samples resulting from these warm-up frames must be thrown
// away, and any decoding errors must be ignored. Once the warm-up is done, the
// decoder should be able operate as normal.
package mpseek
