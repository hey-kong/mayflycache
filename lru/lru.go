package lru

import "container/list"

type LRUCache struct {
	maxBytes  int64 // maximum bytes of memory available
	curBytes  int64 // current bytes of memory in use
	l         *list.List
	m         map[string]*list.Element
	onEvicted func(string, Value) // optional func, called when an Entry is deleted
}

type Entry struct {
	key   string
	value Value
}

func (e Entry) size() int64 {
	return int64(len(e.key)) + int64(e.value.Size())
}

type Value interface {
	Size() int
}

func NewLRUCache(maxBytes int64, onEvicted func(string, Value)) *LRUCache {
	return &LRUCache{
		maxBytes:  maxBytes,
		curBytes:  0,
		l:         list.New(),
		m:         make(map[string]*list.Element),
		onEvicted: onEvicted,
	}
}

func (lru *LRUCache) Get(key string) (value Value, done bool) {
	if e, ok := lru.m[key]; ok {
		// If the cached value is used, it will be moved to the end of the list
		lru.l.MoveToBack(e)

		kv := e.Value.(*Entry)
		value, done = kv.value, true
	}
	return
}

func (lru *LRUCache) Set(key string, value Value) {
	if e, ok := lru.m[key]; ok {
		// If the value already exists, move it to the end of the list and update the value
		lru.l.MoveToBack(e)
		kv := e.Value.(*Entry)
		lru.curBytes -= kv.size()
		kv.value = value
		lru.curBytes += kv.size()
	} else {
		// Otherwise, add a new entry
		e := lru.l.PushBack(&Entry{key, value})
		lru.m[key] = e
		lru.curBytes += e.Value.(*Entry).size()
	}

	// Delete the entry at the head of the list,
	// until the memory occupied is less than the maxBytes.
	for lru.maxBytes != 0 && lru.maxBytes < lru.curBytes {
		lru.Remove()
	}
}

func (lru *LRUCache) Remove() {
	e := lru.l.Front()
	if e != nil {
		kv := e.Value.(*Entry)
		// Removed the entry at the head of the list,
		// the head of the list must be the LRU (least recently used) entry.
		lru.l.Remove(e)
		delete(lru.m, kv.key)
		lru.curBytes -= kv.size()
		// Call callback function
		if lru.onEvicted != nil {
			lru.onEvicted(kv.key, kv.value)
		}
	}
}

// Len returns how many key-value entries are currently cached.
func (lru *LRUCache) Len() int {
	return lru.l.Len()
}
