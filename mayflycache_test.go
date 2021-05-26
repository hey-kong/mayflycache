package main

import (
	"fmt"
	"log"
	"testing"
)

var info = map[string]string{
	"Name":  "Iggie Wang",
	"Age":   "21",
	"Hobby": "League of Legends",
}

func TestGroup(t *testing.T) {
	queryCount := make(map[string]int)
	g := NewGroup("info", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("Load", key, "from database")
			if value, ok := info[key]; ok {
				queryCount[key] += 1
				return []byte(value), nil
			} else {
				return nil, fmt.Errorf("%s not exists", key)
			}
		},
	))

	for k, v := range info {
		if value, err := g.Get(k); err != nil || value.String() != v {
			log.Fatal("First get test failed")
		}

		if _, err := g.Get(k); err != nil || queryCount[k] > 1 {
			log.Fatal("Second get test failed")
		}
	}

	if _, err := g.Get("Unknown"); err == nil {
		log.Fatal("Third get test failed")
	}
}
