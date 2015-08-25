package protocol

import (
	"fmt"

	"github.com/mjibson/mog/codec"
	"golang.org/x/oauth2"
)

type Protocol struct {
	*Params
	OAuth       *oauth2.Config
	newInstance func([]string, *oauth2.Token) (Instance, error)
}

type Params struct {
	Params   []string `json:",omitempty"`
	OAuthURL string   `json:",omitempty"`
}

type Instance interface {
	// Key returns a unique identifier for the instance.
	Key() string
	// List returns the list of available songs, possibly cached.
	List() (SongList, error)
	// Refresh forces an update of the song list.
	Refresh() (SongList, error)
	// Info returns information about one song.
	Info(codec.ID) (*codec.SongInfo, error)
	// GetSong returns a playable song.
	GetSong(codec.ID) (codec.Song, error)
}

type SongList map[codec.ID]*codec.SongInfo

func (p *Protocol) NewInstance(params []string, token *oauth2.Token) (Instance, error) {
	return p.newInstance(params, token)
}

var protocols = make(map[string]*Protocol)

func Register(name string, params []string, newInstance func([]string, *oauth2.Token) (Instance, error)) {
	protocols[name] = &Protocol{
		Params: &Params{
			Params: params,
		},
		newInstance: newInstance,
	}
}

func RegisterOAuth(name string, config *oauth2.Config, newInstance func([]string, *oauth2.Token) (Instance, error)) {
	protocols[name] = &Protocol{
		Params: &Params{
			OAuthURL: config.AuthCodeURL(""),
		},
		OAuth:       config,
		newInstance: newInstance,
	}
}

func ByName(name string) (*Protocol, error) {
	p, ok := protocols[name]
	if !ok {
		return nil, fmt.Errorf("unknown protocol")
	}
	return p, nil
}

func Get() map[string]Params {
	m := make(map[string]Params)
	for n, p := range protocols {
		m[n] = *p.Params
	}
	return m
}
