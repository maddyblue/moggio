package crc16

import (
	"io"
	"testing"
)

type test struct {
	want uint16
	in   string
}

var golden = []test{
	{0x0000, ""},
	{0x8145, "a"},
	{0xC749, "ab"},
	{0xCADB, "abc"},
	{0x58E7, "abcd"},
	{0x678D, "abcde"},
	{0x0D05, "abcdef"},
	{0x047C, "abcdefg"},
	{0x7D68, "abcdefgh"},
	{0x6878, "abcdefghi"},
	{0xF80F, "abcdefghij"},
	{0x0F8E, "Discard medicine more than two years old."},
	{0xE149, "He who has a shady past knows that nice guys finish last."},
	{0x02B7, "I wouldn't marry him with a ten foot pole."},
	{0x7F6A, "Free! Free!/A trip/to Mars/for 900/empty jars/Burma Shave"},
	{0x28BD, "The days of the digital watch are numbered.  -Tom Stoppard"},
	{0x7C55, "Nepal premier won't resign."},
	{0xC92B, "For every action there is an equal and opposite government program."},
	{0x3E41, "His money is twice tainted: 'taint yours and 'taint mine."},
	{0xDA56, "There is no reason for any individual to have a computer in their home. -Ken Olsen, 1977"},
	{0x7F66, "It's a tiny change to the code and not completely disgusting. - Bob Manchek"},
	{0x2A00, "size:  a.out:  bad magic"},
	{0x25B2, "The major problem is with sendmail.  -Mark Horton"},
	{0xBD71, "Give me a rock, paper and scissors and I will move the world.  CCFestoon"},
	{0x8596, "If the enemy is within range, then so are you."},
	{0x74A2, "It's well we cannot hear the screams/That we create in others' dreams."},
	{0x0D73, "You remind me of a TV show, but that's all right: I watch it anyway."},
	{0xEE65, "C is as portable as Stonehedge!!"},
	{0xA94E, "Even if I could be Shakespeare, I think I should still choose to be Faraday. - A. Huxley"},
	{0x0B98, "The fugacity of a constituent in a mixture of gases at a given temperature is proportional to its mole fraction.  Lewis-Randall Rule"},
	{0xF560, "How can you write a big system without C++?  -Paul Glick"},
	{0x60AE, "The quick brown fox jumps over the lazy dog"},
}

func TestCrc16IBM(t *testing.T) {
	for _, g := range golden {
		h := NewIBM()
		io.WriteString(h, g.in)
		got := h.Sum16()
		if got != g.want {
			t.Errorf("IBM(%q); expected 0x%04X, got 0x%04X.", g.in, g.want, got)
		}
	}
}

func BenchmarkNewIBM(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewIBM()
	}
}

func BenchmarkCrc16_1K(b *testing.B) {
	benchmarkCrc16(b, 1024)
}

func BenchmarkCrc16_2K(b *testing.B) {
	benchmarkCrc16(b, 2*1024)
}

func BenchmarkCrc16_4K(b *testing.B) {
	benchmarkCrc16(b, 4*1024)
}

func BenchmarkCrc16_8K(b *testing.B) {
	benchmarkCrc16(b, 8*1024)
}

func BenchmarkCrc16_16K(b *testing.B) {
	benchmarkCrc16(b, 16*1024)
}

func benchmarkCrc16(b *testing.B, count int64) {
	b.SetBytes(count)
	data := make([]byte, count)
	for i := range data {
		data[i] = byte(i)
	}
	h := NewIBM()
	in := make([]byte, 0, h.Size())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Reset()
		h.Write(data)
		h.Sum(in)
	}
}
