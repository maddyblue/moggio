package drive

import (
	"fmt"
	"io"
	"log"

	"net/http"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/google-api-go-client/drive/v2"

	"github.com/mjibson/mog/codec"
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
