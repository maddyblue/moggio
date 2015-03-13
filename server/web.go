package server

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/mjibson/mog/_third_party/github.com/julienschmidt/httprouter"
	"github.com/mjibson/mog/_third_party/golang.org/x/net/websocket"
	"github.com/mjibson/mog/protocol"
)

var indexHTML []byte

func (srv *Server) GetMux(devMode bool) *http.ServeMux {
	var err error
	webFS := FS(devMode)
	if devMode {
		log.Println("using local web assets")
	}
	index, err := webFS.Open("/static/index.html")
	if err != nil {
		log.Fatal(err)
	}
	indexHTML, err = ioutil.ReadAll(index)
	if err != nil {
		log.Fatal(err)
	}
	router := httprouter.New()
	router.GET("/api/cmd/:cmd", JSON(srv.Cmd))
	router.GET("/api/data/:type", JSON(srv.Data))
	router.GET("/api/oauth/:protocol", srv.OAuth)
	router.POST("/api/cmd/:cmd", JSON(srv.Cmd))
	router.POST("/api/queue/change", JSON(srv.QueueChange))
	router.POST("/api/playlist/change/:playlist", JSON(srv.PlaylistChange))
	router.POST("/api/protocol/add", JSON(srv.ProtocolAdd))
	router.POST("/api/protocol/remove", JSON(srv.ProtocolRemove))
	router.POST("/api/protocol/refresh", JSON(srv.ProtocolRefresh))
	mux := http.NewServeMux()
	mux.Handle("/static/", http.FileServer(webFS))
	mux.HandleFunc("/", Index)
	mux.Handle("/api/", router)
	mux.Handle("/ws/", websocket.Handler(srv.WebSocket))
	return mux
}

// ListenAndServe listens on the TCP network address addr and then calls
// Serve to handle requests on incoming connections.
func (srv *Server) ListenAndServe(addr string, devMode bool) error {
	mux := srv.GetMux(devMode)
	log.Println("mog: listening on", addr)
	return http.ListenAndServe(addr, mux)
}

func Index(w http.ResponseWriter, r *http.Request) {
	w.Write(indexHTML)
}

func serveError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func JSON(h func(url.Values, httprouter.Params) (interface{}, error)) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			serveError(w, err)
			return
		}
		d, err := h(r.Form, ps)
		if err != nil {
			serveError(w, err)
			return
		}
		if d == nil {
			return
		}
		b, err := json.Marshal(d)
		if err != nil {
			serveError(w, err)
			return
		}
		w.Header().Add("Content-Type", "application/json")
		w.Write(b)
	}
}

func (srv *Server) OAuth(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	done := make(chan error)
	srv.ch <- cmdAddOAuth{
		name: ps.ByName("protocol"),
		r:    r,
		done: done,
	}
	err := <-done
	if err != nil {
		serveError(w, err)
		return
	}
	http.Redirect(w, r, "/", http.StatusFound)
}

func (srv *Server) Data(form url.Values, ps httprouter.Params) (interface{}, error) {
	return srv.makeWaitData(waitType(ps.ByName("type")))
}

func (srv *Server) Cmd(form url.Values, ps httprouter.Params) (interface{}, error) {
	switch cmd := ps.ByName("cmd"); cmd {
	case "play":
		srv.ch <- cmdPlay
	case "stop":
		srv.ch <- cmdStop
	case "next":
		srv.ch <- cmdNext
	case "prev":
		srv.ch <- cmdPrev
	case "pause":
		srv.ch <- cmdPause
	case "play_idx":
		i, err := strconv.Atoi(form.Get("idx"))
		if err != nil {
			return nil, err
		}
		srv.ch <- cmdPlayIdx(i)
	case "random":
		srv.ch <- cmdRandom
	case "repeat":
		srv.ch <- cmdRepeat
	case "seek":
		d, err := time.ParseDuration(form.Get("pos"))
		if err != nil {
			return nil, err
		}
		srv.ch <- cmdSeek(d)
	default:
		return nil, fmt.Errorf("unknown command: %v", cmd)
	}
	return nil, nil
}

func (srv *Server) QueueChange(form url.Values, ps httprouter.Params) (interface{}, error) {
	srv.ch <- cmdQueueChange(form)
	return nil, nil
}

func (srv *Server) PlaylistChange(form url.Values, ps httprouter.Params) (interface{}, error) {
	srv.ch <- cmdPlaylistChange{
		form: form,
		name: ps.ByName("playlist"),
	}
	return nil, nil
}

func (srv *Server) ProtocolRefresh(form url.Values, ps httprouter.Params) (interface{}, error) {
	return nil, srv.protocolRefresh(form.Get("protocol"), form.Get("key"), false)
}

func (srv *Server) ProtocolAdd(form url.Values, ps httprouter.Params) (interface{}, error) {
	p := form.Get("protocol")
	prot, err := protocol.ByName(p)
	if err != nil {
		return nil, err
	}
	inst, err := prot.NewInstance(form["params"], nil)
	if err != nil {
		return nil, err
	}
	srv.Protocols[p][inst.Key()] = inst
	err = srv.protocolRefresh(p, inst.Key(), false)
	if err != nil {
		delete(srv.Protocols[p], inst.Key())
		return nil, err
	}
	return nil, nil
}

func (srv *Server) ProtocolRemove(form url.Values, ps httprouter.Params) (interface{}, error) {
	p := form.Get("protocol")
	k := form.Get("key")
	prots, ok := srv.Protocols[p]
	if !ok {
		return nil, fmt.Errorf("unknown protocol: %v", p)
	}
	srv.ch <- cmdProtocolRemove{
		protocol: p,
		key:      k,
		prots:    prots,
	}
	return nil, nil
}
