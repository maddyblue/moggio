package drive

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"

	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/protocol"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
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
	protocol.RegisterOAuth("drive", config, New, reflect.TypeOf(&Drive{}))
}

func (d *Drive) getService() (*drive.Service, *http.Client, error) {
	c := config.Client(oauth2.NoContext, d.Token)
	s, err := drive.New(c)
	return s, c, err
}

type Drive struct {
	Token *oauth2.Token
	Name  string
	Files map[string]*drive.File
	Songs protocol.SongList
}

func New(params []string, token *oauth2.Token) (protocol.Instance, error) {
	if token == nil {
		return nil, fmt.Errorf("expected oauth token")
	}
	d := &Drive{
		Token: token,
	}
	service, _, err := d.getService()
	if err != nil {
		return nil, err
	}
	about, err := service.About.Get().Do()
	if err != nil {
		return nil, err
	}
	d.Name = about.Name
	return d, nil
}

func (d *Drive) Key() string {
	return d.Name
}

func (d *Drive) Info(id codec.ID) (*codec.SongInfo, error) {
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

func (d *Drive) GetSong(id codec.ID) (codec.Song, error) {
	path, child := id.Pop()
	f := d.Files[path]
	if f == nil {
		return nil, fmt.Errorf("missing %v", path)
	}
	return codec.ByExtensionID(f.FileExtension, child, d.reader(path))
}

func (d *Drive) reader(id string) codec.Reader {
	return func() (io.ReadCloser, int64, error) {
		log.Println("DRIVE", id)
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
	var ss codec.Songs
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
				info, _ := v.Info()
				if info.Title == "" {
					title := f.Title
					if len(ss) != 1 {
						title += fmt.Sprintf(":%v", i)
					}
					info.Title = title
				}
				songs[codec.NewID(f.Id, string(i))] = &info
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
