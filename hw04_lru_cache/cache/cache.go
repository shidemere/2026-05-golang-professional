// Package cache provides custom LRU cache.
package cache

import (
	"sync"

	"github.com/shidemere/2026-05-golang-professional/hw04_lru_cache/list"
)

// Key is the alias to the string, is our key in cache.
type Key string

// Cache is structure for holding objects in memory.
type Cache interface {
	Set(key Key, value any) bool
	Get(key Key) (any, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    list.List
	items    map[Key]*list.Item
	mu       sync.RWMutex
}

// Clear allows us to clear cache.
func (l *lruCache) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	for _, v := range l.items {
		l.queue.Remove(v)
	}
	clear(l.items)
}

// Get allows us to get any value from cache.
func (l *lruCache) Get(key Key) (any, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	i, ok := l.items[key]
	if !ok {
		return nil, false
	}

	l.queue.MoveToFront(i)
	return i.Value, true
}

// Set allows add any value to the cache by key and return flag of existing.
func (l *lruCache) Set(key Key, value any) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	i, ok := l.items[key]
	if !ok {
		// need to add if doesn't exist
		// if max -> we need push out last item from the list
		return addToCache(l, key, value)
	}
	// update value if it's nedd
	updateInCache(i, value, l)
	return true
}

func updateInCache(i *list.Item, value any, l *lruCache) {
	if i.Value != value {
		l.queue.MoveToFront(i)
		i.Value = value
	} else {
		l.queue.MoveToFront(i)
	}
}

func addToCache(l *lruCache, key Key, value any) bool {
	if l.capacity == l.queue.Len() {
		l.queue.Remove(l.queue.Back())
		delete(l.items, key)
		val := l.queue.PushFront(value)
		l.items[key] = val
	} else {
		val := l.queue.PushFront(value)
		l.items[key] = val
	}
	return false
}

// NewCache creates a new cache.
func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    list.NewList(),
		items:    make(map[Key]*list.Item, capacity),
	}
}
