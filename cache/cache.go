package cache

// cache/cache.go

import (
	"bufio"
	"container/list"
	"encoding/json"
	"fmt"
	"opencache/internal"
	"os"
	"sync"
	"time"
)

type entry struct {
	key       interface{}
	value     interface{}
	expiresAt *time.Time
}

type OpenCache struct {
	Cache      map[interface{}]*list.Element
	Mu         sync.Mutex
	LRU_deque  *list.List
	Capacity   int
	Persistent bool
	LogPath    *string
}

// constructor
func NewOpenCache(capacity int, persistent bool, logPath string) *OpenCache {
	if capacity < 1 {
		capacity = 1
	}
	return &OpenCache{
		Cache:      make(map[interface{}]*list.Element),
		LRU_deque:  list.New(),
		Capacity:   capacity,
		Persistent: persistent,
		LogPath:    &logPath,
	}
}

func (kv *OpenCache) Get(key interface{}) (interface{}, bool) {
	kv.Mu.Lock()
	defer kv.Mu.Unlock()

	// validate key first
	if err := internal.ValidateKey(key); err != nil {
		return nil, false
	}

	// update deque ordering
	if elem, ok := kv.Cache[key]; ok {

		entry := elem.Value.(*entry)
		if entry.expiresAt != nil && time.Now().After(*entry.expiresAt) {
			// if it is past expiration date, then remove from cache and deque
			delete(kv.Cache, entry.key)
			kv.LRU_deque.Remove(elem)
			return nil, false
		}

		kv.LRU_deque.MoveToFront(elem)
		return entry.value, true
	}
	return nil, false
}

func (kv *OpenCache) Set(key interface{}, value interface{}, ttl_duration *time.Duration) bool {
	kv.Mu.Lock()
	defer kv.Mu.Unlock()

	// validate key first
	if err := internal.ValidateKey(key); err != nil {
		return false
	}

	if elem, ok := kv.Cache[key]; ok {
		ent := elem.Value.(*entry)
		ent.value = value
		if ttl_duration != nil {
			exp := time.Now().Add(*ttl_duration)
			ent.expiresAt = &exp
		} else {
			ent.expiresAt = nil
		}
		kv.LRU_deque.MoveToFront(elem)

		// persist update
		if kv.Persistent {
			kv.AppendToLog(makeLogEntry("SET", key, value, ttl_duration))
		}

		return true
	}

	// reached capacity for deque
	if kv.LRU_deque.Len() >= kv.Capacity {

		// right end of deque
		back := kv.LRU_deque.Back()
		if back != nil {
			evicted := back.Value.(*entry)
			delete(kv.Cache, evicted.key)
			kv.LRU_deque.Remove(back)
		}
	}

	var expPtr *time.Time
	if ttl_duration != nil {
		exp := time.Now().Add(*ttl_duration)
		expPtr = &exp
	}

	elem := kv.LRU_deque.PushFront(&entry{
		key:       key,
		value:     value,
		expiresAt: expPtr,
	})
	kv.Cache[key] = elem

	// persist update
	if kv.Persistent {
		kv.AppendToLog(makeLogEntry("SET", key, value, ttl_duration))
	}

	return true
}

func (kv *OpenCache) Delete(key interface{}) bool {
	kv.Mu.Lock()
	defer kv.Mu.Unlock()

	// validate key first
	if err := internal.ValidateKey(key); err != nil {
		return false
	}

	if elem, ok := kv.Cache[key]; ok {
		kv.LRU_deque.Remove(elem)
		delete(kv.Cache, key)

		// persist delete
		if kv.Persistent {
			kv.AppendToLog(makeLogEntry("DELETE", key, nil, nil))
		}
		return true
	}

	return false
}

func (kv *OpenCache) Len() int {
	kv.Mu.Lock()
	defer kv.Mu.Unlock()

	return len(kv.Cache)
}

func (kv *OpenCache) Log() {
	fmt.Println("\nSTART LOG-")
	for k, elem := range kv.Cache {
		e := elem.Value.(*entry).value
		fmt.Printf("Key: %v, Value: %v\n", k, e)
	}
	fmt.Println("END LOG-")
}

// persistence struct + funcs
type LogEntry struct {
	Op    string      `json:"op"`
	Key   string      `json:"key"`
	Value interface{} `json:"value,omitempty"`
	TTLms int64       `json:"ttl_ms,omitempty"`
}

// for every set and delete op, we append to aof file
func (kv *OpenCache) AppendToLog(entry LogEntry) {
	if len(*kv.LogPath) == 0 {
		tmp := "appendonly.aof"
		kv.LogPath = &tmp
	}

	f, err := os.OpenFile(*kv.LogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("log error:", err)
		return
	}
	defer f.Close()

	data, _ := json.Marshal(entry)
	f.Write(append(data, '\n'))
}

// this is just to reset and repopulate the cache given the logfile
func (kv *OpenCache) ReplayLog(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Bug fix-Temporarily disable persistence during replay
	prevPersistent := kv.Persistent
	kv.Persistent = false
	defer func() { kv.Persistent = prevPersistent }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		var entry LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}

		switch entry.Op {
		case "SET":
			var ttl *time.Duration
			if entry.TTLms > 0 {
				t := time.Duration(entry.TTLms) * time.Millisecond
				ttl = &t
			}
			kv.Set(entry.Key, entry.Value, ttl)
		case "DELETE":
			kv.Delete(entry.Key)
		}
	}
	return scanner.Err()
}

func makeLogEntry(op string, key interface{}, value interface{}, ttl *time.Duration) LogEntry {
	strKey := key.(string)
	var ttlms int64
	if ttl != nil {
		ttlms = ttl.Milliseconds()
	}
	return LogEntry{
		Op:    op,
		Key:   strKey,
		Value: value,
		TTLms: ttlms,
	}
}
