package lru

import (
	"reflect"
	"testing"
)

type String string

func (s String) Size() int {
	return len(s)
}

func TestGet(t *testing.T) {
	lru := NewLRUCache(int64(0), nil)
	lru.Set("key1", String("1234"))

	if v, ok := lru.Get("key1"); !ok || string(v.(String)) != "1234" {
		t.Fatalf("cache hit key1=1234 failed")
	}
	if _, ok := lru.Get("key2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestSet(t *testing.T) {
	lru := NewLRUCache(int64(0), nil)
	lru.Set("key", String("1"))
	lru.Set("key", String("111"))

	if lru.curBytes != int64(len("key")+len("111")) {
		t.Fatal("expected 6 but got", lru.curBytes)
	}
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "key1", "key2", "k3"
	v1, v2, v3 := "value1", "value2", "v3"
	cap := len(k1 + k2 + v1 + v2)
	lru := NewLRUCache(int64(cap), nil)
	lru.Set(k1, String(v1))
	lru.Set(k2, String(v2))
	lru.Set(k3, String(v3))

	if _, ok := lru.Get("key1"); ok || lru.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}

func TestOnEvicted(t *testing.T) {
	keys := make([]string, 0)
	callback := func(key string, value Value) {
		keys = append(keys, key)
	}
	lru := NewLRUCache(int64(10), callback)
	lru.Set("key1", String("123456"))
	lru.Set("k2", String("k2"))
	lru.Set("k3", String("k3"))
	lru.Set("k4", String("k4"))

	if expect := []string{"key1", "k2"}; !reflect.DeepEqual(expect, []string{"key1", "k2"}) {
		t.Fatalf("Call OnEvicted failed, expect keys equals to %s", expect)
	}
}
