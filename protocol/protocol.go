package protocol

import (
	"fmt"

	"github.com/golang/oauth2"

	"github.com/mjibson/mog/codec"
)

type Protocol struct {
	Params   []string
	OAuth    *oauth2.Config
	OAuthURL string

	List List
}

type List func(*Instance) (SongList, error)

type ProtocolParams struct {
	Params   []string `json:",omitempty"`
	OAuthURL string   `json:",omitempty"`
}

func (p *Protocol) ProtocolParams() ProtocolParams {
	return ProtocolParams{
		Params:   p.Params[:],
		OAuthURL: p.OAuthURL,
	}
}

type Instance struct {
	Params     []string      `json:",omitempty"`
	OAuthToken *oauth2.Token `json:",omitempty"`
}

type SongList map[string]codec.Song

var protocols = make(map[string]*Protocol)

func Register(name string, params []string, list List) {
	protocols[name] = &Protocol{
		Params: params,
		List:   list,
	}
}

func RegisterOAuth(name string, config *oauth2.Config, list List) {
	protocols[name] = &Protocol{
		OAuth:    config,
		OAuthURL: config.AuthCodeURL("", "", ""),
		List:     list,
	}
}

func ByName(name string) (*Protocol, error) {
	p, ok := protocols[name]
	if !ok {
		return nil, fmt.Errorf("unknown protocol")
	}
	return p, nil
}

func Get() map[string]ProtocolParams {
	m := make(map[string]ProtocolParams)
	for n, p := range protocols {
		m[n] = p.ProtocolParams()
	}
	return m
}

func ListSongs(name string, inst *Instance) (SongList, error) {
	p, ok := protocols[name]
	if !ok {
		return nil, fmt.Errorf("unknown protocol")
	}
	return p.List(inst)
}
