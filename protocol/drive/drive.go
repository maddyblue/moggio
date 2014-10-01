package drive

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/google-api-go-client/drive/v2"

	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/codec/mpa"
	"github.com/mjibson/mog/protocol"
)

var config *oauth.Config

func Init(clientID, clientSecret, redirect string) {
	config = &oauth.Config{
		ClientId:     clientID,
		ClientSecret: clientSecret,
		Scope:        drive.DriveReadonlyScope,
		AuthURL:      "https://accounts.google.com/o/oauth2/auth",
		TokenURL:     "https://accounts.google.com/o/oauth2/token",
		RedirectURL:  redirect + "drive",
	}
	protocol.RegisterOAuth("drive", config, List)
}

func getService(token *oauth.Token) (*drive.Service, *http.Client, error) {
	t := &oauth.Transport{
		Config: config,
		Token:  token,
	}
	c := t.Client()
	s, err := drive.New(c)
	return s, c, err
}

func List(inst *protocol.Instance) (protocol.SongList, error) {
	service, client, err := getService(inst.OAuthToken)
	if err != nil {
		return nil, err
	}
	songs := make(protocol.SongList)
	var ss []codec.Song
	var nextPage string
	for {
		fl, err := service.Files.
			List().
			PageToken(nextPage).
			Fields("nextPageToken", "items(id,fileExtension,fileSize,title)").
			Do()
		if err != nil {
			return nil, err
		}
		nextPage = fl.NextPageToken
		for _, f := range fl.Items {
			f := f
			ss, _, err = codec.ByExtension(f.FileExtension, func() (io.ReadCloser, int64, error) {
				fmt.Println("DRIVE DOWNLOAD", f.Title)
				file, err := service.Files.Get(f.Id).Fields("downloadUrl").Do()
				if err != nil {
					return nil, 0, err
				}
				resp, err := client.Get(file.DownloadUrl)
				if err != nil {
					return nil, 0, err
				}
				if resp.StatusCode != 200 {
					resp.Body.Close()
					return nil, 0, fmt.Errorf(resp.Status)
				}
				return resp.Body, file.FileSize, nil
			})
			if err != nil {
				continue
			}
			for i, v := range ss {
				id := fmt.Sprintf("%v-%v", i, f.Id)
				songs[id] = v
			}
		}
		if nextPage == "" {
			break
		}
	}
	return songs, err
}

type Song struct {
	*GMusic
	*Track
	m *mpa.Song
}

func (s *Song) Init() (sampleRate, channels int, err error) {
	s.m = &mpa.Song{
		Reader: func() (io.ReadCloser, int64, error) {
			log.Println("gmusic get stream", s.Track.ID)
			r, err := s.GMusic.GetStream(s.Track.ID)
			if err != nil {
				return nil, 0, err
			}
			return r.Body, 0, nil
		},
	}
	return s.m.Init()
}

func (s *Song) Play(n int) ([]float32, error) {
	return s.m.Play(n)
}

func (s *Song) Info() (codec.SongInfo, error) {
	duration, _ := strconv.Atoi(s.DurationMillis)
	return codec.SongInfo{
		Time:   time.Duration(duration) * time.Millisecond,
		Artist: s.Artist,
		Title:  s.Title,
		Album:  s.Album,
		Track:  s.TrackNumber,
	}, nil
}

func (s *Song) Close() {
	s.m.Close()
	s.m = nil
}

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

func Login(un, pw, deviceID string) (*GMusic, error) {
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
	gm := GMusic{
		DeviceID: deviceID,
	}
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
	var data ListPlaylists
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
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
	var data ListPlaylistEntries
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
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
	var data ListTracks
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
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
