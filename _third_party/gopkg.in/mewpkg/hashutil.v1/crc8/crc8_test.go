package crc8

import (
	"io"
	"testing"
)

type test struct {
	want uint8
	in   string
}

var golden = []test{
	{0x00, ""},
	{0x20, "a"},
	{0xC9, "ab"},
	{0x5F, "abc"},
	{0xA1, "abcd"},
	{0x52, "abcde"},
	{0x8C, "abcdef"},
	{0x9F, "abcdefg"},
	{0xCB, "abcdefgh"},
	{0x67, "abcdefghi"},
	{0x23, "abcdefghij"},
	{0x56, "Discard medicine more than two years old."},
	{0x6B, "He who has a shady past knows that nice guys finish last."},
	{0x70, "I wouldn't marry him with a ten foot pole."},
	{0x8F, "Free! Free!/A trip/to Mars/for 900/empty jars/Burma Shave"},
	{0x48, "The days of the digital watch are numbered.  -Tom Stoppard"},
	{0x5E, "Nepal premier won't resign."},
	{0x3C, "For every action there is an equal and opposite government program."},
	{0xA8, "His money is twice tainted: 'taint yours and 'taint mine."},
	{0x46, "There is no reason for any individual to have a computer in their home. -Ken Olsen, 1977"},
	{0xC7, "It's a tiny change to the code and not completely disgusting. - Bob Manchek"},
	{0x31, "size:  a.out:  bad magic"},
	{0xB6, "The major problem is with sendmail.  -Mark Horton"},
	{0x7D, "Give me a rock, paper and scissors and I will move the world.  CCFestoon"},
	{0xDC, "If the enemy is within range, then so are you."},
	{0x13, "It's well we cannot hear the screams/That we create in others' dreams."},
	{0x96, "You remind me of a TV show, but that's all right: I watch it anyway."},
	{0x96, "C is as portable as Stonehedge!!"},
	{0x3C, "Even if I could be Shakespeare, I think I should still choose to be Faraday. - A. Huxley"},
	{0xEE, "The fugacity of a constituent in a mixture of gases at a given temperature is proportional to its mole fraction.  Lewis-Randall Rule"},
	{0x33, "How can you write a big system without C++?  -Paul Glick"},
	{0xC1, "The quick brown fox jumps over the lazy dog"},
}

func TestCrc8ATM(t *testing.T) {
	for _, g := range golden {
		h := NewATM()
		io.WriteString(h, g.in)
		got := h.Sum8()
		if got != g.want {
			t.Errorf("ATM(%q); expected 0x%02X, got 0x%02X.", g.in, g.want, got)
		}
	}
}

func BenchmarkNewATM(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewATM()
	}
}

func BenchmarkCrc8_1K(b *testing.B) {
	benchmarkCrc8(b, 1024)
}

func BenchmarkCrc8_2K(b *testing.B) {
	benchmarkCrc8(b, 2*1024)
}

func BenchmarkCrc8_4K(b *testing.B) {
	benchmarkCrc8(b, 4*1024)
}

func BenchmarkCrc8_8K(b *testing.B) {
	benchmarkCrc8(b, 8*1024)
}

func BenchmarkCrc8_16K(b *testing.B) {
	benchmarkCrc8(b, 16*1024)
}

func benchmarkCrc8(b *testing.B, count int64) {
	b.SetBytes(count)
	data := make([]byte, count)
	for i := range data {
		data[i] = byte(i)
	}
	h := NewATM()
	in := make([]byte, 0, h.Size())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Reset()
		h.Write(data)
		h.Sum(in)
	}
}
