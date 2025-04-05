# OpenCache

OpenCache is a lightweight, in-memory key-value store implemented in Go. It supports optional TTL-based expiration and persistence using append-only logs.

## Features

- O(1) `Get`, `Set`, `Delete`, and `Update` operations
- Optional TTL (Time-to-Live) expiration per key
- Optional persistence via append-only log (AOF)
- Least Recently Used (LRU) cache algorithm

## API
- Set(key (must be a comparable type), value interface{}, ttl *time.Duration)
- Get(key interface{}) (interface{}, bool)
- Delete(key interface{}) bool
- ReplayLog(filename string) error

