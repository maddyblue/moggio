package mp3

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

func TestMp3(t *testing.T) {
	f, err := os.Open("test.mp3")
	if err != nil {
		t.Fatal(err)
	}
	m, err := New(f)
	if err != nil {
		t.Fatal(err)
	}
	for m.Scan() {
		f := m.Frame()
		b, _ := json.MarshalIndent(f, "", "  ")
		fmt.Println(string(b))
		break
	}
	if err := m.Err(); err != nil {
		t.Fatal(err)
	}
}
