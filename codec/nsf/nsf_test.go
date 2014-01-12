package nsf

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/mjibson/pulsego"
)

func TestNSF(t *testing.T) {
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

	d := time.Second * 10
	samples := n.Play(d)
	fmt.Println("samples", len(samples))
	ns := 0
	for _, s := range samples {
		if s != 0 {
			ns++
		}
	}
	fmt.Println("ns", ns)
	if ns > 0 {
		pa := pulsego.NewPulseMainLoop()
		defer pa.Dispose()
		pa.Start()

		ctx := pa.NewContext("default", 0)
		if ctx == nil {
			t.Fatal("Failed to create a new context")
		}
		defer ctx.Dispose()
		st := ctx.NewStream("default", &pulsego.PulseSampleSpec{
			Format: pulsego.SAMPLE_FLOAT32LE, Rate: SampleRate, Channels: 1})
		if st == nil {
			t.Fatal("Failed to create a new stream")
		}
		defer st.Dispose()
		st.ConnectToSink()
		st.Write(samples, pulsego.SEEK_RELATIVE)
		time.Sleep(d)
	}
}
