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

type cacheItem struct {
	key   Key
	value any
}

// Clear allows us to clear cache.
func (l *lruCache) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.queue = list.NewList()
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
	return i.Value.(cacheItem).value, true
}

// Set allows add any value to the cache by key and return flag of existing.
func (l *lruCache) Set(key Key, value any) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	i, ok := l.items[key]
	if !ok {
		// need to add if doesn't exist
		// if max -> we need push out last item from the list
		addToCache(l, key, value)
		return false
	}
	// update value if it's nedd
	updateInCache(i, value, l)
	return true
}

func updateInCache(i *list.Item, value any, l *lruCache) {
	item := i.Value.(cacheItem)
	if item.value != value {
		item.value = value
		i.Value = item
	}
	l.queue.MoveToFront(i)
}

func addToCache(l *lruCache, key Key, value any) {
	if l.capacity == 0 {
		return
	}

	if l.capacity == l.queue.Len() {
		evicted := l.queue.Back()
		delete(l.items, evicted.Value.(cacheItem).key)
		l.queue.Remove(evicted)
	}

	val := l.queue.PushFront(cacheItem{key: key, value: value})
	l.items[key] = val
}

// NewCache creates a new cache.
func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    list.NewList(),
		items:    make(map[Key]*list.Item, capacity),
	}
}
