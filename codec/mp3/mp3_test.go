package mp3

import (
	"bytes"
	"os"
	"testing"
)

func TestMp3(t *testing.T) {
	f, err := os.Open("he_44khz.bit")
	if err != nil {
		t.Fatal(err)
	}
	m := New(f)
	m.Sequence()
}

func TestHuffmanTable(t *testing.T) {
	table := huffmanTables[29]
	r := newBitReader(bytes.NewBuffer([]byte{
		0xfd,
	}))
	expected := [][2]byte{
		{0, 0},
		{0, 1},
	}
	for _, e := range expected {
		got := table.tree.Decode(r)
		if got != e {
			t.Fatal("expected", e, "got", got)
		}
	}
}

func TestHuffman(t *testing.T) {
	l := []huffmanPair{
		{[]byte{1}, [2]byte{0, 0}},
		{[]byte{0, 0, 1}, [2]byte{0, 1}},
		{[]byte{0, 1}, [2]byte{1, 0}},
		{[]byte{0, 0, 0}, [2]byte{1, 1}},
	}
	h, err := newHuffmanTree(l)
	if err != nil {
		t.Fatal(err)
	}
	type Test struct {
		input  []byte
		output [][2]byte
	}
	tests := []Test{
		{[]byte{0xf0}, [][2]byte{
			{0, 0},
			{0, 0},
			{0, 0},
			{0, 0},
			{1, 1},
		}},
	}
	for _, test := range tests {
		r := newBitReader(bytes.NewBuffer(test.input))
		for i, v := range test.output {
			got := h.Decode(r)
			if err := r.Err(); err != nil {
				t.Fatal(err)
			}
			if got != v {
				t.Fatalf("%v: got %v, expected %v", i, got, v)
			}
		}
	}
}
