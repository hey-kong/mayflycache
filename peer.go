package main

import pb "github.com/hey-kong/mayflycache/mayflycachepb"

// A PeerPicker interface uses a key to find the PeerGetter
// according to the consistent hash algorithm.
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// A PeerGetter interface is used to get the cached value from the group.
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}
