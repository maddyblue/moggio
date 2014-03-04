package gmusic

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	clientLoginURL = "https://www.google.com/accounts/ClientLogin"
	serviceName    = "sj"
	sourceName     = "gmusicapi-3.1.1-dev"
	sjURL          = "https://www.googleapis.com/sj/v1.1/"
)

type GMusic struct {
	auth string
}

func Login(un, pw string) (*GMusic, error) {
	values := url.Values{}
	values.Add("accountType", "HOSTED_OR_GOOGLE")
	values.Add("Email", un)
	values.Add("Passwd", pw)
	values.Add("service", serviceName)
	values.Add("source", sourceName)
	resp, err := http.PostForm(clientLoginURL, values)
	if err != nil {
		return nil, err
	}
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}
	var gm GMusic
	for _, line := range strings.Fields(string(b)) {
		sp := strings.SplitN(line, "=", 2)
		if len(sp) < 2 {
			continue
		}
		switch sp[0] {
		case "Auth":
			gm.auth = sp[1]
		case "Error":
			return nil, fmt.Errorf("gmusic login: %s", sp[1])
		}
	}
	if gm.auth == "" {
		return nil, fmt.Errorf("gmusic: %s", resp.Status)
	}
	return &gm, nil
}

func (g *GMusic) Request(method, path string) (*http.Response, error) {
	req, err := http.NewRequest(method, sjURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("GoogleLogin auth=%s", g.auth))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gmusic: %s", resp.Status)
	}
	return resp, nil
}

func (g *GMusic) ListPlaylists() ([]*Playlist, error) {
	r, err := g.Request("POST", "playlistfeed")
	if err != nil {
		return nil, err
	}
	body, _ := ioutil.ReadAll(r.Body)
	var data ListPlaylists
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	return data.Data.Items, nil
}

type ListPlaylists struct {
	Data struct {
		Items []*Playlist `json:"items"`
	} `json:"data"`
	Kind string `json:"kind"`
}

type Playlist struct {
	AccessControlled      bool   `json:"accessControlled"`
	CreationTimestamp     string `json:"creationTimestamp"`
	Deleted               bool   `json:"deleted"`
	ID                    string `json:"id"`
	Kind                  string `json:"kind"`
	LastModifiedTimestamp string `json:"lastModifiedTimestamp"`
	Name                  string `json:"name"`
	OwnerName             string `json:"ownerName"`
	OwnerProfilePhotoUrl  string `json:"ownerProfilePhotoUrl"`
	RecentTimestamp       string `json:"recentTimestamp"`
	ShareToken            string `json:"shareToken"`
	Type                  string `json:"type"`
}
