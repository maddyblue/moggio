package mp3

import (
	"bytes"
	"math"
	"os"
	"testing"
)

func TestMp3(t *testing.T) {
	f, err := os.Open("he_44khz.bit")
	if err != nil {
		t.Fatal(err)
	}
	m, err := New(f)
	if err != nil {
		t.Fatal(err)
	}
	n := 1152
	s := m.Play(n)
	if len(s) != n {
		t.Fatalf("bad read len, got %d, expected %d", len(s), n)
	}
	for i, v := range s {
		if !Float64Equal(float64(v), float64(he_44khz_frame1[i])) {
			t.Fatalf("%v: expected %v, got %v\n", i, he_44khz_frame1[i], v)
		}
	}
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

const (
	closeFactor = 1e-8
)

// PrettyClose returns true if the slices a and b are very close, else false.
func PrettyClose(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}

	for i, c := range a {
		if !Float64Equal(c, b[i]) {
			return false
		}
	}
	return true
}

// Float64Equal returns true if a and b are very close, else false.
func Float64Equal(a, b float64) bool {
	return math.Abs(a-b) <= closeFactor || math.Abs(1-a/b) <= closeFactor
}
