package hzycache

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"hzycache/consistenthash"
)

const (
	defaultBasePath = "/_hzycache/"
	defaultReplicas = 50
)

// HTTPPool implements PeerPicker for a pool of HTTP peers.
type HttpPool struct {
	// this peer's base URL, e.g. "https://example.net:8000"
	self        string
	basePath    string
	mu          sync.RWMutex // guards peers and httpGetters
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter
}

func NewHttpPool(self string) *HttpPool {
	return &HttpPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HttpPool) Log(format string, v ...interface{}) {
	log.Printf("[server is %s], %s", p.self, fmt.Sprintf(format, v...))
}

// ServeHTTP handle all http requests
func (p *HttpPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.basePath) {
		panic("HttpPool serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s %s", r.Method, r.URL.Path)
	// /<basepath>/<groupname>/<key> required
	parts := strings.SplitN(r.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	keyName := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(keyName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.b)
}

// Set updates the pool's list of peers.
func (p *HttpPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.peers = consistenthash.New(defaultReplicas, nil)
	p.peers.Add(peers...)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{baseURL: peer + p.basePath}
	}
}

// PickPeer picks a peer according to key
func (p *HttpPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

type httpGetter struct {
	baseURL string
}

// Get get from http
func (h httpGetter) Get(group, key string) ([]byte, error) {
	url := fmt.Sprintf("%v%v/%v",
		h.baseURL, group, key)

	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server return %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body:%w", err)
	}

	return bytes, nil
}

// interface check
var _ PeerGetter = (*httpGetter)(nil)
