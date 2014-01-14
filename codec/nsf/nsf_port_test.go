package nsf

import (
	"os"
	"testing"
	"time"

	"code.google.com/p/portaudio-go/portaudio"
)

func TestPort(t *testing.T) {
	f, err := os.Open("mm3.nsf")
	if err != nil {
		t.Fatal(err)
	}
	n, err := ReadNSF(f)
	if err != nil {
		t.Fatal(err)
	}
	if n.LoadAddr != 0x8000 || n.InitAddr != 0x8003 || n.PlayAddr != 0x8000 {
		t.Error("bad addresses")
	}
	n.Init(1)

	portaudio.Initialize()
	defer portaudio.Terminate()
	var samples []float32
	stream, err := portaudio.OpenDefaultStream(0, 1, float64(n.SampleRate), 1024, func(out []float32) {
		for len(samples) < len(out) {
			s := n.Play(time.Second / 10)
			samples = append(samples, s...)
		}
		for i := range out {
			out[i] = samples[0]
			samples = samples[1:]
		}
	})
	if err != nil {
		t.Fatal(err)
	}
	defer stream.Close()
	if err := stream.Start(); err != nil {
		t.Fatal(err)
	}
	time.Sleep(time.Second * 500)
	if err := stream.Stop(); err != nil {
		t.Fatal(err)
	}
}
