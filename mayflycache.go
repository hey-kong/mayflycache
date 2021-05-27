package main

import (
	"fmt"
	"log"
	"sync"

	pb "github.com/hey-kong/mayflycache/mayflycachepb"
)

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

type Group struct {
	name      string
	mainCache *SafeCache
	getter    Getter
	peers     PeerPicker
	once      Once
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// NewGroup initializes all fields except PeerPicker,
// which needs to call RegisterPeers.
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("Nil Getter")
	}

	mu.Lock()
	defer mu.Unlock()

	g := &Group{
		name:      name,
		mainCache: &SafeCache{maxBytes: cacheBytes},
		getter:    getter,
	}
	groups[name] = g
	return g
}

// RegisterPeers registers PeerPicker of the Group.
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeers called more than once")
	}
	g.peers = peers
}

// GetGroup returns the group.
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// It tries to get the cached data from its mainCache;
// If not, call g.load to use Getter or get data from peer node.
func (g *Group) Get(key string) (Chunk, error) {
	// Null key is handled here to prevent cache penetration
	if key == "" {
		return Chunk{}, fmt.Errorf("key is required")
	}
	// Try to get a cached chunk, and return it if you get it
	if v, ok := g.mainCache.Get(key); ok {
		log.Println("Cache Hit")
		return v, nil
	}
	// Otherwise, load the data into the cache
	return g.load(key)
}

// If its peers is nil，call getLocally to get;
// Else call peers.PickPeer to get peer node, and call getFromPeer to get data from remote.
func (g *Group) load(key string) (value Chunk, err error) {
	tmpValue, err := g.once.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("Failed to get from peer:", err)
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return tmpValue.(Chunk), nil
	}
	return
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (Chunk, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return Chunk{}, err
	}
	return NewChunk(res.Value), nil
}

func (g *Group) getLocally(key string) (value Chunk, err error) {
	// Call getter to get data
	bytes, err := g.getter.Get(key)
	if err != nil {
		return
	}
	// Save the data to the chunk and cache it
	value = NewChunk(bytes)
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value Chunk) {
	g.mainCache.Set(key, value)
}
