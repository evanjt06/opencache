# OpenCache

OpenCache is a lightweight, in-memory key-value store implemented in Go. It supports optional TTL-based expiration and persistence using append-only logs.

## Features

- O(1) `Get`, `Set`, `Delete` operations
- Optional TTL (Time-to-Live) expiration per key
- Optional persistence via append-only log (AOF) file
- Least Recently Used (LRU) cache algorithm

## API
- Set(key (must be a comparable type), value interface{}, ttl *time.Duration)
- Get(key interface{}) (interface{}, bool)
- Delete(key interface{}) bool
- ReplayLog(filename string) error

## Motivations
I made this as a basic exercise to make my own (KV) cache system like Redis. Opencache is an in-memory key-value (KV) store / cache that implements the Least Recently Used (LRU) eviction algorithm to free up space when the cache's preset capacity is overflowed. I also added a Time to live (TTL) feature for key expiration for cache records. all operations for read, update, and delete are O(1) b/c of the double-ended queue and cache ops. validation for the cache keys have also been added so that only comparable elements can be keys. Opencache is also thread-safe because I used mutex, so it is multi-client safe. Added basic persistence with AOF log files.
<br>
<br>
If used in production, note that Opencache makes your API/service stateful. the general consensus is that REST APIs / gRPC services should be stateless due to scaling reasoning b/c of horizontal scaling. So do not use opencache when 1) there is a need for distributed cache between services 2) long-term storage 3) pub/sub 4) super big datasets like >GB.
<br>
<br>
ideally, Opencache should be used on the application layer of your system for 1) non-mission critical data that can be recomputed from your DB or just reduce expensive operations 2) session or token storage (NOTE: only when session data can be stored elsewhere like in your db or redis) 3) rate limiting purposes w/ TTL exp
