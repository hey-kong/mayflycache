package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/hey-kong/mayflycache/consistenthash"
)

const (
	defaultBasePath = "/mayflycache/"
	defaultReplicas = 50
)

// A HTTPPool represents the HTTP server structure, it implements
// the PeerPicker and ServeHTTP interface to provide the cached value.
type HTTPPool struct {
	self        string // for log output and verifying the service
	basePath    string // equal to defaultBasePath
	mu          sync.Mutex
	peers       *consistenthash.Map    // consistent hash
	httpGetters map[string]*httpGetter // map node name to httpGetter
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (hp *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s]%s", hp.self, fmt.Sprintf(format, v...))
}

func (hp *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hp.Log("%s %s", r.Method, r.URL.Path)

	// Serve only '/mayflycache/*' requests
	if !strings.HasPrefix(r.URL.Path, hp.basePath) {
		http.Error(w, "HTTPPool serving unexpected path:"+r.URL.Path, http.StatusBadRequest)
		return
	}

	// The format of the request is '/<basePath>/<groupName>/<key>',
	// parts is []string{<groupName>, <key>} here.
	parts := strings.SplitN(r.URL.Path[len(hp.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName, key := parts[0], parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}

	value, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(value.ByteSlice())
}

// Set delays the assignment of peers and httpGetters.
func (hp *HTTPPool) Set(peers ...string) {
	hp.mu.Lock()
	defer hp.mu.Unlock()

	hp.peers = consistenthash.New(defaultReplicas, nil)
	hp.peers.Set(peers...)
	hp.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		hp.httpGetters[peer] = &httpGetter{
			baseURL: peer + hp.basePath,
		}
	}
}

// PickPeer implements PeerPicker interface for HTTPPool to return the httpGetter according to the key.
func (hp *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	hp.mu.Lock()
	defer hp.mu.Unlock()
	if peer := hp.peers.Get(key); peer != "" && peer != hp.self {
		hp.Log("Pick peer %s", peer)
		return hp.httpGetters[peer], true
	}
	return nil, false
}

// httpGetter is an implementation of PeerGetter on HTTP protocol.
type httpGetter struct {
	baseURL string
}

// Get uses baseURL, group and key to splice request URL,
// and calls http.Get to get data from a group.
func (hp *httpGetter) Get(group, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		hp.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Server returned: %v\n", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("error when reading response body: %v", err)
	}

	return bytes, nil
}
