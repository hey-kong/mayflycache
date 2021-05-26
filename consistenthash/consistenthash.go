package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash           // hash function
	replicas int            // record how many virtual nodes a real node corresponds to
	keys     []int          // hash ring
	hashMap  map[int]string // the mapping between virtual nodes and real nodes
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		hash:     fn,
		replicas: replicas,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Get uses the key to calculate the corresponding node name
// in the hashMap according to the consistent hash algorithm.
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// Find the first index greater than or equal to the hash in the hash ring
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	// If not found, sort.Search will return len(m.keys),
	// we need to set it to 0.
	return m.hashMap[m.keys[idx%len(m.keys)]]
}

// Set adds virtual nodes to the hash ring.
func (m *Map) Set(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}
