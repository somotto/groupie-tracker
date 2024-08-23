package cache

import (
	"sync" // safely handle concurrent access to the cache.
	"time"
)

/*
This file defines a basic cache system that stores key-value pairs with expiration times.
It uses a sync.RWMutex to handle concurrent access safely, allowing multiple readers or a single writer at a time.
 It provides methods to add items to the cache (Set) and retrieve items from the cache (Get), handling expiration and concurrency concerns.
 Cache allows many readers to access the data at the same time,
 which can be faster than using a regular lock (like sync.Mutex)
 that would block everyone, even other readers.
*/

// This shows one item in the cache
type CacheItem struct {
	Value      interface{}
	Expiration time.Time
}

type Cache struct {
	items map[string]CacheItem
	mu    sync.RWMutex
}

func NewCache() *Cache {
	return &Cache{
		items: make(map[string]CacheItem),
	}
}

func (c *Cache) Set(key string, value interface{}, duration time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = CacheItem{
		Value:      value,
		Expiration: time.Now().Add(duration),
	}
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, found := c.items[key]
	if !found {
		return nil, false
	}
	if time.Now().After(item.Expiration) {
		delete(c.items, key)
		return nil, false
	}
	return item.Value, true
}
