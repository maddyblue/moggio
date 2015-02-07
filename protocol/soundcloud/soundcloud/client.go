package soundcloud

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/mjibson/mog/_third_party/golang.org/x/oauth2"

	"github.com/mjibson/mog/_third_party/code.google.com/p/google-api-go-client/googleapi"
)

const basePath = "https://api.soundcloud.com/"

func New(client *http.Client, token *oauth2.Token) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	if token == nil {
		return nil, errors.New("token is nil")
	}
	base, err := url.Parse(basePath)
	if err != nil {
		return nil, err
	}
	s := &Service{client: client, token: token, BasePath: base}
	return s, nil
}

type Service struct {
	client   *http.Client
	token    *oauth2.Token
	BasePath *url.URL
}

func (s *Service) Me() *MeCall {
	c := &MeCall{s: s, opt_: make(map[string]interface{})}
	return c
}

type MeCall struct {
	s    *Service
	opt_ map[string]interface{}
}

func (c *MeCall) Do() (*Me, error) {
	params := make(url.Values)
	params.Set("oauth_token", c.s.token.AccessToken)
	urls, err := c.s.BasePath.Parse("me.json")
	if err != nil {
		return nil, err
	}
	urls.RawQuery = params.Encode()
	req, _ := http.NewRequest("GET", urls.String(), nil)
	res, err := c.s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	var ret *Me
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

type Me struct {
	AvatarURL             string      `json:"avatar_url"`
	City                  interface{} `json:"city"`
	Country               interface{} `json:"country"`
	Description           interface{} `json:"description"`
	DiscogsName           interface{} `json:"discogs_name"`
	FirstName             string      `json:"first_name"`
	FollowersCount        int64       `json:"followers_count"`
	FollowingsCount       int64       `json:"followings_count"`
	FullName              string      `json:"full_name"`
	ID                    int64       `json:"id"`
	Kind                  string      `json:"kind"`
	LastModified          string      `json:"last_modified"`
	LastName              string      `json:"last_name"`
	MyspaceName           interface{} `json:"myspace_name"`
	Online                bool        `json:"online"`
	Permalink             string      `json:"permalink"`
	PermalinkURL          string      `json:"permalink_url"`
	Plan                  string      `json:"plan"`
	PlaylistCount         int64       `json:"playlist_count"`
	PrimaryEmailConfirmed bool        `json:"primary_email_confirmed"`
	PrivatePlaylistsCount int64       `json:"private_playlists_count"`
	PrivateTracksCount    int64       `json:"private_tracks_count"`
	PublicFavoritesCount  int64       `json:"public_favorites_count"`
	Quota                 struct {
		UnlimitedUploadQuota bool  `json:"unlimited_upload_quota"`
		UploadSecondsLeft    int64 `json:"upload_seconds_left"`
		UploadSecondsUsed    int64 `json:"upload_seconds_used"`
	} `json:"quota"`
	Subscriptions     []interface{} `json:"subscriptions"`
	TrackCount        int64         `json:"track_count"`
	UploadSecondsLeft int64         `json:"upload_seconds_left"`
	Uri               string        `json:"uri"`
	Username          string        `json:"username"`
	Website           interface{}   `json:"website"`
	WebsiteTitle      interface{}   `json:"website_title"`
}

func (s *Service) Favorites() *FavoritesCall {
	c := &FavoritesCall{s: s, opt_: make(map[string]interface{})}
	return c
}

type FavoritesCall struct {
	s    *Service
	opt_ map[string]interface{}
}

func (c *FavoritesCall) Do() ([]*Favorite, error) {
	params := make(url.Values)
	params.Set("oauth_token", c.s.token.AccessToken)
	urls, err := c.s.BasePath.Parse("me/favorites.json")
	if err != nil {
		return nil, err
	}
	urls.RawQuery = params.Encode()
	req, _ := http.NewRequest("GET", urls.String(), nil)
	res, err := c.s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	var ret []*Favorite
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

type Favorite struct {
	ArtworkURL          string      `json:"artwork_url"`
	AttachmentsUri      string      `json:"attachments_uri"`
	Bpm                 interface{} `json:"bpm"`
	CommentCount        int64       `json:"comment_count"`
	Commentable         bool        `json:"commentable"`
	CreatedAt           string      `json:"created_at"`
	Description         string      `json:"description"`
	DownloadCount       int64       `json:"download_count"`
	Downloadable        bool        `json:"downloadable"`
	Duration            int64       `json:"duration"`
	EmbeddableBy        string      `json:"embeddable_by"`
	FavoritingsCount    int64       `json:"favoritings_count"`
	Genre               string      `json:"genre"`
	ID                  int64       `json:"id"`
	Isrc                string      `json:"isrc"`
	KeySignature        string      `json:"key_signature"`
	Kind                string      `json:"kind"`
	LabelID             interface{} `json:"label_id"`
	LabelName           string      `json:"label_name"`
	LastModified        string      `json:"last_modified"`
	License             string      `json:"license"`
	OriginalContentSize int64       `json:"original_content_size"`
	OriginalFormat      string      `json:"original_format"`
	Permalink           string      `json:"permalink"`
	PermalinkURL        string      `json:"permalink_url"`
	PlaybackCount       int64       `json:"playback_count"`
	Policy              string      `json:"policy"`
	PurchaseTitle       interface{} `json:"purchase_title"`
	PurchaseURL         interface{} `json:"purchase_url"`
	Release             string      `json:"release"`
	ReleaseDay          interface{} `json:"release_day"`
	ReleaseMonth        interface{} `json:"release_month"`
	ReleaseYear         interface{} `json:"release_year"`
	Sharing             string      `json:"sharing"`
	State               string      `json:"state"`
	StreamURL           string      `json:"stream_url"`
	Streamable          bool        `json:"streamable"`
	TagList             string      `json:"tag_list"`
	Title               string      `json:"title"`
	TrackType           string      `json:"track_type"`
	Uri                 string      `json:"uri"`
	User                struct {
		AvatarURL    string `json:"avatar_url"`
		ID           int64  `json:"id"`
		Kind         string `json:"kind"`
		LastModified string `json:"last_modified"`
		Permalink    string `json:"permalink"`
		PermalinkURL string `json:"permalink_url"`
		Uri          string `json:"uri"`
		Username     string `json:"username"`
	} `json:"user"`
	UserFavorite      bool   `json:"user_favorite"`
	UserID            int64  `json:"user_id"`
	UserPlaybackCount int64  `json:"user_playback_count"`
	VideoURL          string `json:"video_url"`
	WaveformURL       string `json:"waveform_url"`
}
