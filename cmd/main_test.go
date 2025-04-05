package main

import (
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/evanjt06/opencache/cache"
)

func TestSetAndGet(t *testing.T) {
	c := cache.NewOpenCache(2, false, "")

	c.Set("a", 1, nil)
	c.Set("b", 2, nil)

	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Errorf("Expected 'a' to be 1, got %v", v)
	}

	if v, ok := c.Get("b"); !ok || v != 2 {
		t.Errorf("Expected 'b' to be 2, got %v", v)
	}

	c.Log()
}

func TestLRUEviction(t *testing.T) {
	c := cache.NewOpenCache(2, false, "")

	c.Set("a", 1, nil)
	c.Set("b", 2, nil)
	c.Get("a")         // Access 'a' → now 'b' is LRU
	c.Set("c", 3, nil) // Evicts 'b'

	if _, ok := c.Get("b"); ok {
		t.Error("Expected 'b' to be evicted")
	}

	if _, ok := c.Get("a"); !ok {
		t.Error("Expected 'a' to still be present")
	}

	if _, ok := c.Get("c"); !ok {
		t.Error("Expected 'c' to be present")
	}
	c.Log()
}

func TestTTLExpiration(t *testing.T) {
	c := cache.NewOpenCache(2, false, "")

	ttl := 1 * time.Second
	c.Set("x", "expiring", &ttl)

	time.Sleep(1500 * time.Millisecond)

	if _, ok := c.Get("x"); ok {
		t.Error("Expected 'x' to have expired")
	}

	c.Log()
}

func TestDelete(t *testing.T) {
	c := cache.NewOpenCache(1, false, "")

	c.Set("foo", "bar", nil)
	ok := c.Delete("foo")
	if !ok {
		t.Error("Expected Delete to return true")
	}

	if _, ok := c.Get("foo"); ok {
		t.Error("Expected 'foo' to be deleted")
	}
	c.Log()
}

func TestConcurrentAccess(t *testing.T) {
	c := cache.NewOpenCache(100, false, "")
	var wg sync.WaitGroup

	setAndGet := func(key string, val int) {
		defer wg.Done()
		c.Set(key, val, nil)
		v, ok := c.Get(key)
		if !ok || v != val {
			t.Errorf("Expected %v, got %v", val, v)
		}
	}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go setAndGet(fmt.Sprintf("key-%d", i), i)
	}

	wg.Wait()

	c.Log()
}

func TestPersistentCache(t *testing.T) {
	logFile := "test_appendonly.aof"
	os.Remove(logFile) // clean up previous log

	// 1. Create persistent cache and write some data
	c := cache.NewOpenCache(10, true, logFile)

	c.Delete("hi")
	c.Set("user", "evan", nil)
	ttl := 2 * time.Second
	c.Set("session", "abc123", &ttl)
	c.Delete("user")
	c.Log()

	// 2. Reconstruct from log
	c2 := cache.NewOpenCache(10, true, logFile)
	if err := c2.ReplayLog(logFile); err != nil {
		t.Fatalf("ReplayLog failed: %v", err)
	}
	c2.Log()
}
