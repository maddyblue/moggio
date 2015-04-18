package stream

import (
	"bufio"
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/mjibson/mog/_third_party/golang.org/x/oauth2"
	"github.com/mjibson/mog/codec"
	"github.com/mjibson/mog/codec/mpa"
	"github.com/mjibson/mog/protocol"
)

func init() {
	protocol.Register("stream", []string{"URL"}, New)
	gob.Register(new(Stream))
}

func New(params []string, token *oauth2.Token) (protocol.Instance, error) {
	if len(params) != 1 {
		return nil, fmt.Errorf("expected one parameter")
	}
	u, name := tryPLS(params[0])
	if name == "" {
		name = params[0]
	}
	s := Stream{
		Orig: params[0],
		URL:  u,
		Name: name,
	}
	resp, err := s.get()
	if err != nil {
		return nil, err
	}
	resp.Body.Close()
	return &s, nil
}

// tryPLS checks if u is a URL to a .pls file. If it is, it returns the
// first File entry of as target and first Title entry as name.
func tryPLS(u string) (target, name string) {
	target = u
	resp, err := http.Get(u)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	sc := bufio.NewScanner(io.LimitReader(resp.Body, 1024))
	i := 0
	for sc.Scan() {
		if i > 5 {
			break
		}
		i++
		sp := strings.SplitN(sc.Text(), "=", 2)
		if len(sp) != 2 {
			continue
		}
		if strings.HasPrefix(sp[0], "File") {
			_, err := url.Parse(sp[1])
			if err != nil {
				return
			}
			target = sp[1]
		} else if strings.HasPrefix(sp[0], "Title") {
			name = sp[1]
		}
		if target != u && name != "" {
			return
		}
	}
	return
}

type Stream struct {
	Orig string
	URL  string
	Name string
}

type dialer struct {
	*net.Dialer
}

type conn struct {
	net.Conn
	read bool
}

// Read modifies the first line of an ICY stream response,
// if needed, to conform to Go's HTTP version requirements:
// http://golang.org/pkg/net/http/#ParseHTTPVersion.
func (c *conn) Read(b []byte) (n int, err error) {
	if !c.read {
		const headerICY = "ICY"
		const headerHTTP = "HTTP/1.1"
		// Hold 5 bytes because "HTTP/1.1" is 5 bytes longer than "ICY".
		n, err := c.Conn.Read(b[:len(b)+len(headerICY)-len(headerHTTP)])
		if bytes.HasPrefix(b, []byte(headerICY)) {
			copy(b[len(headerHTTP):], b[len(headerICY):])
			copy(b, []byte(headerHTTP))
		}
		c.read = true
		return n, err
	}
	return c.Conn.Read(b)
}

func (d *dialer) Dial(network, address string) (net.Conn, error) {
	c, err := d.Dialer.Dial(network, address)
	cn := conn{
		Conn: c,
	}
	return &cn, err
}

var client = &http.Client{
	Transport: &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		Dial: (&dialer{
			Dialer: &net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			},
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	},
}

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
		Title: s.Name,
	}
}

func (s *Stream) Key() string {
	return s.Orig
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
