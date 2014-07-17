package protocol

import (
	"fmt"

	"github.com/mjibson/mog/codec"
)

type protocol struct {
	params []string
	list   func([]string) (SongList, error)
}

type SongList map[string]codec.Song

var protocols = make(map[string]protocol)

func Register(name string, params []string, list func([]string) (SongList, error)) {
	protocols[name] = protocol{params, list}
}

func Get() map[string][]string {
	m := make(map[string][]string)
	for n, p := range protocols {
		m[n] = p.params
	}
	return m
}

func List(name string, params []string) (SongList, error) {
	p, ok := protocols[name]
	if !ok {
		return nil, fmt.Errorf("unknown protocol")
	}
	return p.list(params)
}
