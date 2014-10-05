package soundcloud

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"code.google.com/p/google-api-go-client/googleapi"

	"github.com/golang/oauth2"
	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/codec/mpa"
	"github.com/mjibson/mog/protocol"
	"github.com/mjibson/mog/protocol/soundcloud/soundcloud"
)

var config *oauth2.Config
var oauthClientID string

func Init(clientID, clientSecret, redirect string) {
	c, err := oauth2.NewConfig(
		&oauth2.Options{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirect + "soundcloud",
			Scopes:       []string{"non-expiring"},
		},
		"https://soundcloud.com/connect",
		"https://api.soundcloud.com/oauth2/token",
	)
	if err != nil {
		log.Fatal(err)
	}
	config = c
	oauthClientID = clientID
	protocol.RegisterOAuth("soundcloud", config, List)
}

func getService(token *oauth2.Token) (*soundcloud.Service, *http.Client, error) {
	t := config.NewTransport()
	t.SetToken(token)
	c := &http.Client{Transport: t}
	s, err := soundcloud.New(c, token)
	return s, c, err
}

func List(inst *protocol.Instance) (protocol.SongList, error) {
	service, client, err := getService(inst.OAuthToken)
	if err != nil {
		return nil, err
	}
	favorites, err := service.Favorites().Do()
	if err != nil {
		return nil, err
	}
	songs := make(protocol.SongList)
	var ss []codec.Song
	for _, f := range favorites {
		f := f
		ss, err = mpa.NewSongs(func() (io.ReadCloser, int64, error) {
			fmt.Println("DROPBOX SOUNDCLOUD", f.Title)
			res, err := client.Get(f.StreamURL + "?client_id=" + oauthClientID)
			if err != nil {
				return nil, 0, err
			}
			if err := googleapi.CheckResponse(res); err != nil {
				return nil, 0, err
			}
			return res.Body, 0, nil
		})
		if err != nil {
			continue
		}
		for i, v := range ss {
			id := fmt.Sprintf("%v-%v", i, f.ID)
			songs[id] = v
		}
	}
	return songs, err
}
