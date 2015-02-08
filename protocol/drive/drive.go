package drive

import (
	"encoding/gob"
	"fmt"
	"io"

	"net/http"

	"github.com/mjibson/mog/_third_party/code.google.com/p/google-api-go-client/drive/v2"

	"github.com/mjibson/mog/_third_party/golang.org/x/oauth2"
	"github.com/mjibson/mog/_third_party/golang.org/x/oauth2/google"
	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/protocol"
)

var config *oauth2.Config

func init() {
	gob.Register(new(Drive))
}

func Init(clientID, clientSecret, redirect string) {
	config = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirect + "drive",
		Scopes:       []string{drive.DriveReadonlyScope},
		Endpoint:     google.Endpoint,
	}
	protocol.RegisterOAuth("drive", config, New)
}

func (d *Drive) getService() (*drive.Service, *http.Client, error) {
	c := config.Client(oauth2.NoContext, d.Token)
	s, err := drive.New(c)
	return s, c, err
}

type Drive struct {
	Token *oauth2.Token
	Files map[string]*drive.File
	Songs protocol.SongList
}

func New(params []string, token *oauth2.Token) (protocol.Instance, error) {
	if token == nil {
		return nil, fmt.Errorf("expected oauth token")
	}
	return &Drive{
		Token: token,
	}, nil
}

func (d *Drive) Key() string {
	return d.Token.AccessToken
}

func (d *Drive) Info(id string) (*codec.SongInfo, error) {
	s := d.Songs[id]
	if s == nil {
		return nil, fmt.Errorf("could not find %v", id)
	}
	return s, nil
}

func (d *Drive) List() (protocol.SongList, error) {
	if len(d.Songs) == 0 {
		return d.Refresh()
	}
	return d.Songs, nil
}

func (d *Drive) GetSong(id string) (codec.Song, error) {
	path, num, err := protocol.ParseID(id)
	if err != nil {
		return nil, err
	}
	f := d.Files[path]
	if f == nil {
		return nil, fmt.Errorf("missing %v", path)
	}
	ss, _, err := codec.ByExtension(f.FileExtension, d.reader(path))
	if err != nil {
		return nil, err
	}
	if len(ss) < num+1 {
		return nil, fmt.Errorf("missing %v", id)
	}
	return ss[num], nil
}

func (d *Drive) reader(id string) codec.Reader {
	return func() (io.ReadCloser, int64, error) {
		fmt.Println("DRIVE", id)
		service, client, err := d.getService()
		if err != nil {
			return nil, 0, err
		}
		file, err := service.Files.Get(id).Fields("downloadUrl").Do()
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
	}
}

func (d *Drive) Refresh() (protocol.SongList, error) {
	service, _, err := d.getService()
	if err != nil {
		return nil, err
	}
	files := make(map[string]*drive.File)
	songs := make(protocol.SongList)
	var nextPage string
	var ss []codec.Song
	for {
		fl, err := service.Files.
			List().
			PageToken(nextPage).
			Fields("nextPageToken", "items(id,fileExtension,fileSize,title)").
			MaxResults(1000).
			Do()
		if err != nil {
			return nil, err
		}
		nextPage = fl.NextPageToken
		for _, f := range fl.Items {
			ss, _, err = codec.ByExtension(f.FileExtension, d.reader(f.Id))
			if err != nil || len(ss) == 0 {
				continue
			}
			files[f.Id] = f
			for i, v := range ss {
				id := fmt.Sprintf("%v-%v", i, f.Id)
				info, _ := v.Info()
				if info.Title == "" {
					title := f.Title
					if len(ss) != 1 {
						title += fmt.Sprintf(":%v", i)
					}
					info.Title = title
				}
				songs[id] = &info
			}
		}
		if nextPage == "" {
			break
		}
	}
	d.Songs = songs
	d.Files = files
	return songs, err
}
