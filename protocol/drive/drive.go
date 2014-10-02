package drive

import (
	"fmt"
	"io"
	"log"

	"net/http"

	"code.google.com/p/google-api-go-client/drive/v2"

	"github.com/golang/oauth2"
	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/protocol"
)

var config *oauth2.Config

func Init(clientID, clientSecret, redirect string) {
	c, err := oauth2.NewConfig(
		&oauth2.Options{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirect + "drive",
			Scopes:       []string{drive.DriveReadonlyScope},
		},
		"https://accounts.google.com/o/oauth2/auth",
		"https://accounts.google.com/o/oauth2/token",
	)
	if err != nil {
		log.Fatal(err)
	}
	config = c
	protocol.RegisterOAuth("drive", config, List)
}

func getService(token *oauth2.Token) (*drive.Service, *http.Client, error) {
	t := config.NewTransport()
	t.SetToken(token)
	c := &http.Client{Transport: t}
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
