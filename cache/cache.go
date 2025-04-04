package cache

// cache/cache.go

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

type entry struct {
	key       interface{}
	value     interface{}
	expiresAt *time.Time
}

type OpenCache struct {
	Cache     map[interface{}]*list.Element
	Mu        sync.Mutex
	LRU_deque *list.List
	Capacity  int
}

// constructor
func NewOpenCache(capacity int) *OpenCache {
	if capacity < 1 {
		capacity = 1
	}
	return &OpenCache{
		Cache:     make(map[interface{}]*list.Element),
		LRU_deque: list.New(),
		Capacity:  capacity,
	}
}

func (kv *OpenCache) Get(key interface{}) (interface{}, bool) {
	kv.Mu.Lock()
	defer kv.Mu.Unlock()

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

func (kv *OpenCache) Set(key interface{}, value interface{}, ttl_duration *time.Duration) {
	kv.Mu.Lock()
	defer kv.Mu.Unlock()

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
		return
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

}

func (kv *OpenCache) Delete(key interface{}) bool {
	kv.Mu.Lock()
	defer kv.Mu.Unlock()

	if elem, ok := kv.Cache[key]; ok {
		kv.LRU_deque.Remove(elem)
		delete(kv.Cache, key)
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
