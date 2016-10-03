package youtube

import (
	"encoding/gob"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"github.com/mjibson/moggio/codec"
	"github.com/mjibson/moggio/protocol"
	"github.com/rylio/ytdl"
	"golang.org/x/oauth2"
)

func init() {
	protocol.Register("youtube", []string{"URL"}, New, reflect.TypeOf(&Youtube{}))
	gob.Register(new(Youtube))
}

func New(params []string, token *oauth2.Token) (protocol.Instance, error) {
	if len(params) != 1 {
		return nil, fmt.Errorf("expected one parameter")
	}
	y := Youtube{
		ID: params[0],
	}
	if _, err := y.Refresh(); err != nil {
		return nil, err
	}
	return &y, nil
}

type Youtube struct {
	ID   string
	info codec.SongInfo
}

func (y *Youtube) Key() string {
	return y.ID
}

func (y *Youtube) getInfo() (*ytdl.VideoInfo, ytdl.Format, error) {
	v, err := ytdl.GetVideoInfo(y.ID)
	if err != nil {
		return nil, ytdl.Format{}, err
	}
	var f ytdl.Format
	const itag = 43 // webm
	for _, f = range v.Formats {
		if f.Itag == itag {
			break
		}
	}
	if f.Itag != itag {
		return nil, ytdl.Format{}, fmt.Errorf("could not find audio-only stream")
	}
	return v, f, nil
}

func (y *Youtube) getURL() (*url.URL, error) {
	v, f, err := y.getInfo()
	if err != nil {
		return nil, err
	}
	return v.GetDownloadURL(f)
}

func (y *Youtube) Refresh() (protocol.SongList, error) {
	v, _, err := y.getInfo()
	if err != nil {
		return nil, err
	}
	y.info = codec.SongInfo{
		Time:   v.Duration,
		Artist: v.Author,
		Title:  v.Title,
	}
	if u := v.GetThumbnailURL(ytdl.ThumbnailQualityDefault); u != nil {
		y.info.ImageURL = u.String()
	}
	return y.List()
}

func (y *Youtube) List() (protocol.SongList, error) {
	return protocol.SongList{
		codec.ID(y.ID): &y.info,
	}, nil
}

func (y *Youtube) Info(codec.ID) (*codec.SongInfo, error) {
	return &y.info, nil
}

func (y *Youtube) GetSong(codec.ID) (codec.Song, error) {
	return webm.NewSong(func() (io.ReadCloser, int64, error) {
		u, err := y.getURL()
		if err != nil {
			return nil, 0, err
		}
		resp, err := http.Get(u.String())
		if err != nil {
			return nil, 0, err
		}
		if resp.StatusCode != 200 {
			b, _ := ioutil.ReadAll(io.LimitReader(resp.Body, 1024))
			resp.Body.Close()
			return nil, 0, fmt.Errorf("%v: %s", resp.Status, b)
		}
		return resp.Body, 0, nil
	})
}
