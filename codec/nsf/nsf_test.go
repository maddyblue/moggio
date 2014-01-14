package nsf

import (
	"os"
	"testing"
	"time"

	"github.com/mjibson/pulsego"
)

func _TestPulseGo(t *testing.T) {
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

	pa := pulsego.NewPulseMainLoop()
	defer pa.Dispose()
	pa.Start()

	ctx := pa.NewContext("default", 0)
	if ctx == nil {
		t.Fatal("Failed to create a new context")
	}
	defer ctx.Dispose()
	st := ctx.NewStream("default", &pulsego.PulseSampleSpec{
		Format: pulsego.SAMPLE_FLOAT32LE, Rate: int(n.SampleRate), Channels: 1})
	if st == nil {
		t.Fatal("Failed to create a new stream")
	}
	defer st.Dispose()
	st.ConnectToSink()

	d := time.Second / 5
	for {
		wait := time.After(d)
		start := time.Now()
		samples := n.Play(d)
		since := time.Since(start)
		println("samples", len(samples), 100*since/d)
		st.Write(samples, pulsego.SEEK_RELATIVE)
		<-wait
	}
	time.Sleep(d)
}
