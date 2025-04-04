package main

import (
	"opencache/cache"
)

func main() {

	c := cache.NewOpenCache(1)

	c.Set("Hi", 10, nil)
	c.Log()

	c.Set("S", 20, nil)
	c.Log()
}
