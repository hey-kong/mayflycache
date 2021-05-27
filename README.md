# mayflycache

mayflycache is a simple implementation of a distributed caching and cache-filling library, with reference to groupcache.

## Features

  * HTTP-based.
  * Least Recently Used (LRU) caching strategy.
  * Using mutex locks for thread safety.
  * Implementing singleflight to prevent cache breakdown.
  * Load balancing using consistent hashing.
  * Using protobuf for inter-node communication.

## Example

You can refer to `main.go` and run `run.sh`.

## Related Links

1. [groupcache](https://github.com/golang/groupcache)
2. [geecache](https://github.com/geektutu/7days-golang/tree/master/gee-cache)
