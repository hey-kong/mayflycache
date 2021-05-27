package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var db = map[string]string{
	"Name":  "Iggie Wang",
	"Age":   "21",
	"Hobby": "League of Legends",
}

func createGroup() *Group {
	return NewGroup("info", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("search key", key, "from db")
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exists", key)
		},
	))
}

func startCacheServer(addr string, addrs []string, group *Group) {
	hp := NewHTTPPool(addr)
	hp.Set(addrs...)
	group.RegisterPeers(hp)
	log.Println("CacheServer is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], hp))
}

func startAPIServer(apiAddr string, group *Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			value, err := group.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(value.ByteSlice())
		},
	))
	log.Println("Frontend Server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "CacheServer port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	cache := createGroup()
	if api {
		apiAddr := "http://localhost:9999"
		go startAPIServer(apiAddr, cache)
	}
	startCacheServer(addrMap[port], addrs, cache)
}
