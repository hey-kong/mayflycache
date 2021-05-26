package main

// A PeerPicker interface uses a key to find the PeerGetter
// according to the consistent hash algorithm.
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// A PeerGetter interface is used to get the cached value from the group.
type PeerGetter interface {
	Get(group, key string) ([]byte, error)
}
