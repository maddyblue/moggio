// +build darwin

package output

import (
	"encoding/binary"
	"fmt"

	"github.com/wlcx/ao"
)

type aod struct {
	d *ao.Device
}

func init() {
	ao.Init()
}

func get(sampleRate, channels int) (Output, error) {
	o := &aod{}
	id, err := ao.DefaultDriver()
	if err != nil {
		panic(err)
		return nil, err
	}
	sf := ao.SampleFormat{
		Channels:  channels,
		Rate:      sampleRate,
		Bits:      32,
		ByteOrder: ao.EndianLittle,
	}
	fmt.Printf("%+v\n", sf)
	options := map[string]string{
		"debug":   "",
		"verbose": "",
	}
	o.d, err = ao.OpenLive(id, &sf, options)
	if err != nil {
		panic(err)
		return nil, err
	}
	return o, nil
}

func (a *aod) Push(samples []float32) {
	err := binary.Write(a.d, binary.LittleEndian, samples)
	if err != nil {
		panic(err)
	}
}

func (a *aod) Stop() {
}

func (a *aod) Start() {
}
