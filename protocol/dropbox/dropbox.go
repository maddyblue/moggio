package dropbox

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/golang/oauth2"
	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/protocol"
	"github.com/mjibson/mog/protocol/dropbox/dropbox"
)

var config *oauth2.Config

func Init(clientID, clientSecret, redirect string) {
	c, err := oauth2.NewConfig(
		&oauth2.Options{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirect + "dropbox",
		},
		"https://www.dropbox.com/1/oauth2/authorize",
		"https://api.dropbox.com/1/oauth2/token",
	)
	if err != nil {
		log.Fatal(err)
	}
	config = c
	protocol.RegisterOAuth("dropbox", config, List)
}

func getService(token *oauth2.Token) (*dropbox.Service, *http.Client, error) {
	t := config.NewTransport()
	t.SetToken(token)
	c := &http.Client{Transport: t}
	s, err := dropbox.New(c)
	return s, c, err
}

func List(inst *protocol.Instance) (protocol.SongList, error) {
	service, _, err := getService(inst.OAuthToken)
	if err != nil {
		return nil, err
	}
	songs := make(protocol.SongList)
	var ss []codec.Song
	dirs := []string{"/Audio"}
	for {
		if len(dirs) == 0 {
			break
		}
		dir := dirs[0]
		dirs = dirs[1:]
		list, err := service.List().Path(dir).Do()
		if err != nil {
			return nil, err
		}
		for _, f := range list.Contents {
			if f.IsDir {
				dirs = append(dirs, f.Path)
				continue
			}
			f := f
			idx := strings.LastIndex(f.Path, ".")
			if idx < 0 {
				log.Println("dropbox: no file extension:", f.Path)
				continue
			}
			ext := f.Path[idx+1:]
			ss, _, err = codec.ByExtension(ext, func() (io.ReadCloser, int64, error) {
				fmt.Println("DROPBOX DOWNLOAD", f.Path)
				file, err := service.Get().Path(f.Path).Do()
				if err != nil {
					return nil, 0, err
				}
				return file, f.Bytes, nil
			})
			if err != nil {
				continue
			}
			for i, v := range ss {
				id := fmt.Sprintf("%v-%v", i, f.Path)
				songs[id] = v
			}
		}
	}
	return songs, err
}
