// Package lazycache provides a concurrency-safe lazy-initialization cache
// keyed by string. It uses double-checked locking to minimize contention:
// a read lock checks the fast path, falling back to a write lock with a
// second check only on cache misses.
package lazycache

import "sync"

// Cache is a concurrency-safe map that lazily creates values on first access.
// The zero value is not usable; create instances with New.
type Cache[T any] struct {
	items map[string]T
	mu    sync.RWMutex
}

// New creates an empty Cache.
func New[T any]() *Cache[T] {
	return &Cache[T]{
		items: make(map[string]T),
	}
}

// GetOrCreate returns the cached value for key, or calls create to produce one
// if the key is not yet present. The create function is called at most once per
// key, even under concurrent access. If create returns an error, nothing is
// cached and the error is returned to all concurrent callers that triggered it.
func (c *Cache[T]) GetOrCreate(key string, create func() (T, error)) (T, error) {
	// Fast path: read lock.
	c.mu.RLock()
	if item, ok := c.items[key]; ok {
		c.mu.RUnlock()
		return item, nil
	}
	c.mu.RUnlock()

	// Slow path: write lock with double-check.
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, ok := c.items[key]; ok {
		return item, nil
	}

	item, err := create()
	if err != nil {
		var zero T
		return zero, err
	}

	c.items[key] = item
	return item, nil
}

// Get returns the cached value for key and whether it was found.
func (c *Cache[T]) Get(key string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, ok := c.items[key]
	return item, ok
}

// Set unconditionally stores a value for key.
func (c *Cache[T]) Set(key string, value T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = value
}

// Delete removes a key from the cache.
func (c *Cache[T]) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}
