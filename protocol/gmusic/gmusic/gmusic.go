package gmusic

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const (
	clientLoginURL          = "https://www.google.com/accounts/ClientLogin"
	googlePlayMusicEndpoint = "https://play.google.com/music"
	serviceName             = "sj"
	sourceName              = "mog"
	sjURL                   = "https://www.googleapis.com/sj/v1.1/"
)

type GMusic struct {
	DeviceID string
	Auth     string
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
	gm := GMusic{}
	for _, line := range strings.Fields(string(b)) {
		sp := strings.SplitN(line, "=", 2)
		if len(sp) < 2 {
			continue
		}
		switch sp[0] {
		case "Auth":
			gm.Auth = sp[1]
		case "Error":
			return nil, fmt.Errorf("gmusic login: %s", sp[1])
		}
	}
	if gm.Auth == "" {
		return nil, fmt.Errorf("gmusic: %s", resp.Status)
	}
	if err := gm.setDeviceID(); err != nil {
		return nil, err
	}
	return &gm, nil
}

func (g *GMusic) request(method, url string, data interface{}, client *http.Client) (*http.Response, error) {
	var buf *bytes.Buffer
	if data != nil {
		buf = new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(data); err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", fmt.Sprintf("GoogleLogin auth=%s", g.Auth))
	req.Header.Add("Content-Type", "application/json")
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("gmusic: %s", resp.Status)
	}
	return resp, nil
}

func (g *GMusic) Request(method, path string, data interface{}) (*http.Response, error) {
	return g.request(method, sjURL+path, data, nil)
}

func (g *GMusic) setDeviceID() error {
	req, err := http.NewRequest("HEAD", googlePlayMusicEndpoint+"/listen", nil)
	if err != nil {
		return err
	}
	req.Header.Add("Authorization", "GoogleLogin auth="+g.Auth)
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()
	xt := make(url.Values)
	u, _ := url.Parse(googlePlayMusicEndpoint)
	for _, c := range jar.Cookies(u) {
		if c.Name == "xt" {
			xt.Set("xt", c.Value)
		}
	}
	settings, err := g.settings(xt, client)
	if err != nil {
		return err
	}
	if len(settings.Devices) == 0 || len(settings.Devices[0].ID) < 2 {
		return fmt.Errorf("no valid devices")
	}
	g.DeviceID = settings.Devices[0].ID[2:]
	return nil
}

func (g *GMusic) settings(xtData url.Values, jarClient *http.Client) (*Settings, error) {
	resp, err := g.request("POST", googlePlayMusicEndpoint+"/services/loadsettings?"+xtData.Encode(), nil, jarClient)
	if err != nil {
		return nil, err
	}
	var data SettingsData
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	return &data.Settings, nil
}

type SettingsData struct {
	Settings Settings `json:"settings"`
}

type Settings struct {
	Devices []struct {
		Carrier      string `json:"carrier"`
		Date         int    `json:"date"`
		ID           string `json:"id"`
		LastUsedMs   int    `json:"lastUsedMs"`
		Manufacturer string `json:"manufacturer"`
		Model        string `json:"model"`
		Name         string `json:"name"`
		Type         string `json:"type"`
	} `json:"devices"`
	ExpirationMillis int  `json:"expirationMillis"`
	IsCanceled       bool `json:"isCanceled"`
	IsSubscription   bool `json:"isSubscription"`
	IsTrial          bool `json:"isTrial"`
	Labs             []struct {
		Description string `json:"description"`
		Enabled     bool   `json:"enabled"`
		Name        string `json:"name"`
		Title       string `json:"title"`
	} `json:"labs"`
	MaxTracks              int  `json:"maxTracks"`
	SubscriptionNewsletter bool `json:"subscriptionNewsletter"`
}

func (g *GMusic) ListPlaylists() ([]*Playlist, error) {
	r, err := g.Request("POST", "playlistfeed", nil)
	if err != nil {
		return nil, err
	}
	var data ListPlaylists
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
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

func (g *GMusic) ListPlaylistEntries() ([]*PlaylistEntry, error) {
	r, err := g.Request("POST", "plentryfeed", nil)
	if err != nil {
		return nil, err
	}
	var data ListPlaylistEntries
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		return nil, err
	}
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
	var tracks []*Track
	var next string
	for {
		r, err := g.Request("POST", "trackfeed", struct {
			StartToken string `json:"start-token"`
		}{
			StartToken: next,
		})
		if err != nil {
			return nil, err
		}
		var data ListTracks
		defer r.Body.Close()
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			return nil, err
		}
		tracks = append(tracks, data.Data.Items...)
		next = data.NextPageToken
		if next == "" {
			break
		}
	}
	return tracks, nil
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
	sig, salt := getSignature(songID)
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
	req.Header.Add("Authorization", fmt.Sprintf("GoogleLogin auth=%s", g.Auth))
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

func getSignature(songID string) (sig, salt string) {
	const key = "34ee7983-5ee6-4147-aa86-443ea062abf774493d6a-2a15-43fe-aace-e78566927585\n"
	salt = fmt.Sprint(time.Now().UnixNano() / 1e6)
	mac := hmac.New(sha1.New, []byte(key))
	mac.Write([]byte(songID))
	mac.Write([]byte(salt))
	sig = base64.URLEncoding.EncodeToString(mac.Sum(nil))
	sig = sig[:len(sig)-1]
	return
}
