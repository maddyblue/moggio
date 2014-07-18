package gmusic

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/protocol"
)

func init() {
	protocol.Register("gmusic", []string{"username", "password", "device id"}, List)
}

func List(params []string) (protocol.SongList, error) {
	if len(params) != 3 {
		return nil, fmt.Errorf("bad params")
	}
	g, err := Login(params[0], params[1])
	if err != nil {
		return nil, err
	}
	tracks, err := g.ListTracks()
	if err != nil {
		return nil, err
	}
	songs := make(protocol.SongList)
	for _, t := range tracks {
		songs[t.ID] = &Song{g, t}
	}
	return songs, nil
}

type Song struct {
	*GMusic
	*Track
}

func (s *Song) Init() (sampleRate, channels int, err error) {
	return
}

func (s *Song) Play(n int) ([]float32, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *Song) Info() codec.SongInfo {
	duration, _ := strconv.Atoi(s.DurationMillis)
	return codec.SongInfo{
		Time:       time.Duration(duration) * time.Millisecond,
		Artist:      s.Artist,
		Title:      s.Title,
		Album:      s.Album,
		Track:      s.TrackNumber,
	}
}

func (s *Song) Close() {}

const (
	clientLoginURL = "https://www.google.com/accounts/ClientLogin"
	serviceName    = "sj"
	sourceName     = "gmusicapi-3.1.1-dev"
	sjURL          = "https://www.googleapis.com/sj/v1.1/"
)

type GMusic struct {
	DeviceID string

	Playlists       []*Playlist
	PlaylistEntries []*PlaylistEntry
	Tracks          []*Track

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
	g.Playlists = data.Data.Items
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

func (g *GMusic) ListPlaylistEntries() ([]*PlaylistEntry, error) {
	r, err := g.Request("POST", "plentryfeed")
	if err != nil {
		return nil, err
	}
	body, _ := ioutil.ReadAll(r.Body)
	var data ListPlaylistEntries
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	g.PlaylistEntries = data.Data.Items
	return data.Data.Items, nil
}

type ListPlaylistEntries struct {
	Data struct {
		Items []*PlaylistEntry `json:"items"`
	} `json:"data"`
	Kind          string `json:"kind"`
	NextPageToken string `json:"nextPageToken"`
}

type PlaylistEntry struct {
	AbsolutePosition      string `json:"absolutePosition"`
	ClientId              string `json:"clientId"`
	CreationTimestamp     string `json:"creationTimestamp"`
	Deleted               bool   `json:"deleted"`
	ID                    string `json:"id"`
	Kind                  string `json:"kind"`
	LastModifiedTimestamp string `json:"lastModifiedTimestamp"`
	PlaylistId            string `json:"playlistId"`
	Source                string `json:"source"`
	TrackId               string `json:"trackId"`
}

func (g *GMusic) ListTracks() ([]*Track, error) {
	r, err := g.Request("POST", "trackfeed")
	if err != nil {
		return nil, err
	}
	body, _ := ioutil.ReadAll(r.Body)
	var data ListTracks
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}
	g.Tracks = data.Data.Items
	return data.Data.Items, nil
}

type ListTracks struct {
	Data struct {
		Items []*Track `json:"items"`
	} `json:"data"`
	Kind          string `json:"kind"`
	NextPageToken string `json:"nextPageToken"`
}

type Track struct {
	Album       string `json:"album"`
	AlbumArtRef []struct {
		URL string `json:"url"`
	} `json:"albumArtRef"`
	AlbumArtist  string `json:"albumArtist"`
	AlbumId      string `json:"albumId"`
	Artist       string `json:"artist"`
	ArtistArtRef []struct {
		URL string `json:"url"`
	} `json:"artistArtRef"`
	ArtistId              []string `json:"artistId"`
	ClientId              string   `json:"clientId"`
	CreationTimestamp     string   `json:"creationTimestamp"`
	Deleted               bool     `json:"deleted"`
	DiscNumber            float64  `json:"discNumber"`
	DurationMillis        string   `json:"durationMillis"`
	EstimatedSize         string   `json:"estimatedSize"`
	ID                    string   `json:"id"`
	Kind                  string   `json:"kind"`
	LastModifiedTimestamp string   `json:"lastModifiedTimestamp"`
	Nid                   string   `json:"nid"`
	PlayCount             float64  `json:"playCount"`
	RecentTimestamp       string   `json:"recentTimestamp"`
	StoreId               string   `json:"storeId"`
	Title                 string   `json:"title"`
	TrackNumber           float64  `json:"trackNumber"`
	TrackType             string   `json:"trackType"`
	Year                  float64  `json:"year"`
}

func (g *GMusic) GetStream(songID string) (*http.Response, error) {
	sig, salt := GetSignature(songID)
	v := url.Values{}
	v.Add("opt", "hi")
	v.Add("net", "wifi")
	v.Add("pt", "e")
	v.Add("slt", salt)
	v.Add("sig", sig)
	if strings.HasPrefix(songID, "T") {
		v.Add("mjck", songID)
	} else {
		v.Add("songid", songID)
	}
	u := url.URL{
		Scheme:   "https",
		Host:     "android.clients.google.com",
		Path:     "/music/mplay",
		RawQuery: v.Encode(),
	}
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("GoogleLogin auth=%s", g.auth))
	req.Header.Add("X-Device-ID", g.DeviceID)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gmusic: %s", resp.Status)
	}
	return resp, nil
}

func GetSignature(songID string) (sig, salt string) {
	const key = "34ee7983-5ee6-4147-aa86-443ea062abf774493d6a-2a15-43fe-aace-e78566927585\n"
	salt = fmt.Sprint(time.Now().UnixNano() / 1e6)
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(songID))
	mac.Write([]byte(salt))
	sig = base64.URLEncoding.EncodeToString(mac.Sum(nil))
	sig = sig[:len(sig)-1]
	return
}
