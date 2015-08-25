package dropbox

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"

	"google.golang.org/api/googleapi"
)

const (
	basePath    = "https://api.dropbox.com/1/"
	contentPath = "https://api-content.dropbox.com/1/"
)

func New(client *http.Client) (*Service, error) {
	if client == nil {
		return nil, errors.New("client is nil")
	}
	base, err := url.Parse(basePath)
	if err != nil {
		return nil, err
	}
	content, err := url.Parse(contentPath)
	if err != nil {
		return nil, err
	}
	s := &Service{client: client, BasePath: base, ContentPath: content}
	return s, nil
}

type Service struct {
	client      *http.Client
	BasePath    *url.URL
	ContentPath *url.URL
}

func (s *Service) List() *ListCall {
	c := &ListCall{s: s, opt_: make(map[string]interface{})}
	return c
}

type ListCall struct {
	s    *Service
	path string
	opt_ map[string]interface{}
}

func (c *ListCall) Do() (*List, error) {
	params := make(url.Values)
	urls, err := c.s.BasePath.Parse("metadata/auto/" + c.path)
	if err != nil {
		return nil, err
	}
	urls.RawQuery = params.Encode()
	req, _ := http.NewRequest("GET", urls.String(), nil)
	res, err := c.s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if err := googleapi.CheckResponse(res); err != nil {
		return nil, err
	}
	var ret *List
	if err := json.NewDecoder(res.Body).Decode(&ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (c *ListCall) Path(path string) *ListCall {
	c.path = path
	return c
}

type ListContent struct {
	Bytes       int64  `json:"bytes"`
	Icon        string `json:"icon"`
	IsDir       bool   `json:"is_dir"`
	Modified    string `json:"modified"`
	Path        string `json:"path"`
	Rev         string `json:"rev"`
	Revision    int64  `json:"revision"`
	Root        string `json:"root"`
	Size        string `json:"size"`
	ThumbExists bool   `json:"thumb_exists"`
}

type List struct {
	Bytes       int64          `json:"bytes"`
	Contents    []*ListContent `json:"contents"`
	Hash        string         `json:"hash"`
	Icon        string         `json:"icon"`
	IsDir       bool           `json:"is_dir"`
	Path        string         `json:"path"`
	Root        string         `json:"root"`
	Size        string         `json:"size"`
	ThumbExists bool           `json:"thumb_exists"`
}

func (s *Service) Get() *GetCall {
	c := &GetCall{s: s, opt_: make(map[string]interface{})}
	return c
}

type GetCall struct {
	s    *Service
	opt_ map[string]interface{}
	path string
}

func (c *GetCall) Do() (io.ReadCloser, error) {
	params := make(url.Values)
	urls, err := c.s.ContentPath.Parse("files/auto/" + c.path)
	if err != nil {
		return nil, err
	}
	urls.RawQuery = params.Encode()
	req, _ := http.NewRequest("GET", urls.String(), nil)
	res, err := c.s.client.Do(req)
	if err != nil {
		return nil, err
	}
	if err := googleapi.CheckResponse(res); err != nil {
		res.Body.Close()
		return nil, err
	}
	return res.Body, nil
}

func (c *GetCall) Path(path string) *GetCall {
	c.path = path
	return c
}
