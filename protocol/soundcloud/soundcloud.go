package soundcloud

import (
	"encoding/gob"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mjibson/mog/_third_party/github.com/google/google-api-go-client/googleapi"
	"github.com/mjibson/mog/_third_party/golang.org/x/oauth2"
	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/codec/mpa"
	"github.com/mjibson/mog/protocol"
	"github.com/mjibson/mog/protocol/soundcloud/soundcloud"
)

var config *oauth2.Config
var oauthClientID string

func init() {
	gob.Register(new(Soundcloud))
}

func Init(clientID, clientSecret, redirect string) {
	config := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirect + "soundcloud",
		Scopes:       []string{"non-expiring"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://soundcloud.com/connect",
			TokenURL: "https://api.soundcloud.com/oauth2/token",
		},
	}
	oauthClientID = clientID
	protocol.RegisterOAuth("soundcloud", config, New)
}

func (s *Soundcloud) getService() (*soundcloud.Service, *http.Client, error) {
	c := config.Client(oauth2.NoContext, s.Token)
	svc, err := soundcloud.New(c, s.Token)
	return svc, c, err
}

type Soundcloud struct {
	Token     *oauth2.Token
	Favorites map[codec.ID]*soundcloud.Favorite
}

func New(params []string, token *oauth2.Token) (protocol.Instance, error) {
	if token == nil {
		return nil, fmt.Errorf("expected oauth token")
	}
	return &Soundcloud{
		Token: token,
	}, nil
}

func (s *Soundcloud) Key() string {
	return s.Token.AccessToken
}

func (s *Soundcloud) Info(id codec.ID) (*codec.SongInfo, error) {
	f := s.Favorites[id]
	if f == nil {
		return nil, fmt.Errorf("could not find %v", id)
	}
	return toInfo(f), nil
}

func toInfo(f *soundcloud.Favorite) *codec.SongInfo {
	return &codec.SongInfo{
		Time:     time.Duration(f.Duration) * time.Millisecond,
		Artist:   f.User.Username,
		Title:    f.Title,
		ImageURL: f.ArtworkURL,
	}
}

func (s *Soundcloud) SongList() protocol.SongList {
	m := make(protocol.SongList)
	for k, f := range s.Favorites {
		m[k] = toInfo(f)
	}
	return m
}

func (s *Soundcloud) List() (protocol.SongList, error) {
	if len(s.Favorites) == 0 {
		return s.Refresh()
	}
	return s.SongList(), nil
}

func (s *Soundcloud) GetSong(id codec.ID) (codec.Song, error) {
	fmt.Println("SOUNDCLOUD", id)
	_, client, err := s.getService()
	if err != nil {
		return nil, err
	}
	f := s.Favorites[id]
	if f == nil {
		return nil, fmt.Errorf("bad id: %v", id)
	}
	return mpa.NewSong(func() (io.ReadCloser, int64, error) {
		res, err := client.Get(f.StreamURL + "?client_id=" + oauthClientID)
		if err != nil {
			return nil, 0, err
		}
		if err := googleapi.CheckResponse(res); err != nil {
			return nil, 0, err
		}
		return res.Body, 0, nil
	})
}

func (s *Soundcloud) Refresh() (protocol.SongList, error) {
	service, _, err := s.getService()
	if err != nil {
		return nil, err
	}
	favorites, err := service.Favorites().Do()
	if err != nil {
		return nil, err
	}
	favs := make(map[codec.ID]*soundcloud.Favorite)
	for _, f := range favorites {
		favs[codec.Int64(f.ID)] = f
	}
	s.Favorites = favs
	return s.SongList(), err
}
