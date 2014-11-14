package dropbox

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/golang/oauth2"
	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/protocol"
	"github.com/mjibson/mog/protocol/dropbox/dropbox"
)

var config *oauth2.Config

func init() {
	gob.Register(new(Dropbox))
}

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
	protocol.RegisterOAuth("dropbox", config, New)
}

func (d *Dropbox) getService() (*dropbox.Service, error) {
	t := config.NewTransport()
	t.SetToken(d.Token)
	c := &http.Client{Transport: t}
	s, err := dropbox.New(c)
	return s, err
}

type Dropbox struct {
	Token *oauth2.Token
	Files map[string]*dropbox.ListContent
	Songs protocol.SongList
}

func New(params []string, token *oauth2.Token) (protocol.Instance, error) {
	if token == nil {
		return nil, fmt.Errorf("expected oauth token")
	}
	return &Dropbox{
		Token: token,
	}, nil
}

func (d *Dropbox) Key() string {
	return d.Token.AccessToken
}

func (d *Dropbox) Info(id string) (*codec.SongInfo, error) {
	s := d.Songs[id]
	if s == nil {
		return nil, fmt.Errorf("could not find %v", id)
	}
	return s, nil
}

func (d *Dropbox) List() (protocol.SongList, error) {
	if len(d.Songs) == 0 {
		return d.Refresh()
	}
	return d.Songs, nil
}

func (d *Dropbox) GetSong(id string) (codec.Song, error) {
	path, num, err := protocol.ParseID(id)
	if err != nil {
		return nil, err
	}
	f := d.Files[path]
	if f == nil {
		return nil, fmt.Errorf("missing %v", path)
	}
	ss, _, err := codec.ByExtension(f.Path, d.reader(path, f.Bytes))
	if err != nil {
		return nil, err
	}
	if len(ss) < num+1 {
		return nil, fmt.Errorf("missing %v", id)
	}
	return ss[num], nil
}

func (d *Dropbox) reader(id string, size int64) codec.Reader {
	return func() (io.ReadCloser, int64, error) {
		fmt.Println("DROPBOX ", id)
		service, err := d.getService()
		if err != nil {
			return nil, 0, err
		}
		file, err := service.Get().Path(id).Do()
		if err != nil {
			return nil, 0, err
		}
		return file, size, nil
	}
}

func (d *Dropbox) Refresh() (protocol.SongList, error) {
	service, err := d.getService()
	if err != nil {
		return nil, err
	}
	files := make(map[string]*dropbox.ListContent)
	songs := make(protocol.SongList)
	var ss []codec.Song
	var info codec.SongInfo
	dirs := []string{""}
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
			ss, _, err = codec.ByExtension(f.Path, d.reader(f.Path, f.Bytes))
			if err != nil || len(ss) == 0 {
				continue
			}
			files[f.Path] = f
			for i, v := range ss {
				id := fmt.Sprintf("%v-%v", i, f.Path)
				info, err = v.Info()
				if err != nil {
					continue
				}
				songs[id] = &info
			}
		}
	}
	d.Songs = songs
	d.Files = files
	return songs, err
}
