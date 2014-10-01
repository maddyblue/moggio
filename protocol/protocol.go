package protocol

import (
	"fmt"
	"net/http"

	"code.google.com/p/goauth2/oauth"

	"github.com/mjibson/mog/codec"
)

type Protocol struct {
	Params   []string
	OAuth    *oauth.Config
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

func (p *Protocol) OAuthTransport() *oauth.Transport {
	return &oauth.Transport{Config: p.OAuth}
}

func (p *Protocol) GetOAuthClient(inst *Instance) *http.Client {
	t := p.OAuthTransport()
	t.Token = inst.OAuthToken
	return t.Client()
}

type Instance struct {
	Params     []string     `json:",omitempty"`
	OAuthToken *oauth.Token `json:",omitempty"`
}

type SongList map[string]codec.Song

var protocols = make(map[string]*Protocol)

func Register(name string, params []string, list List) {
	protocols[name] = &Protocol{
		Params: params,
		List:   list,
	}
}

func RegisterOAuth(name string, config *oauth.Config, list List) {
	protocols[name] = &Protocol{
		OAuth:    config,
		OAuthURL: config.AuthCodeURL(""),
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
