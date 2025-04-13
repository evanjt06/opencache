package main

import "github.com/evanjt06/opencache/cache"

func main() {
	c := cache.NewOpenCache(2, false, "")

	c.Set("a", 1, nil)
	c.Set("b", 2, nil)
	c.Get("a")         // Access 'a' → now 'b' is LRU
	c.Set("d", 3, nil) // Evicts 'b'

	c.Print()
}
