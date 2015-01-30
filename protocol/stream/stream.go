package stream

import (
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/codec/mpa"
	"github.com/mjibson/mog/protocol"
	"golang.org/x/oauth2"
)

func init() {
	protocol.Register("stream", []string{"URL"}, New)
	gob.Register(new(Stream))
}

func New(params []string, token *oauth2.Token) (protocol.Instance, error) {
	if len(params) != 1 {
		return nil, fmt.Errorf("expected one parameter")
	}
	s := Stream{
		URL: params[0],
	}
	resp, err := s.get()
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	return &s, nil
}

type Stream struct {
	URL string
}

var client = &http.Client{}

func (s *Stream) get() (*http.Response, error) {
	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		panic(err)
		log.Fatal(err)
	}
	//req.Header.Add("Icy-MetaData", "1")
	log.Println("stream open", req.URL)
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("stream status: %v", resp.Status)
	}
	/*
		metaint, err := strconv.ParseInt(resp.Header.Get("Icy-Metaint"), 10, 64)
		if err != nil {
			return nil, err
		}
	*/

	return resp, nil
}

func (s *Stream) info() *codec.SongInfo {
	return &codec.SongInfo{
		Title: s.URL,
	}
}

func (s *Stream) Key() string {
	return s.URL
}

func (s *Stream) List() (protocol.SongList, error) {
	return protocol.SongList{
		s.URL: s.info(),
	}, nil
}

func (s *Stream) Refresh() (protocol.SongList, error) {
	return s.List()
}

func (s *Stream) Info(string) (*codec.SongInfo, error) {
	return s.info(), nil
}

func (s *Stream) GetSong(string) (codec.Song, error) {
	return mpa.NewSong(s.reader())
}

func (s *Stream) reader() codec.Reader {
	return func() (io.ReadCloser, int64, error) {
		resp, err := s.get()
		if err != nil {
			return nil, 0, err
		}
		return resp.Body, 0, nil
	}
}
