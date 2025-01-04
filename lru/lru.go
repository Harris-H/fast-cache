package lru

import (
	"errors"
	"fast-cache/internal"
)

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback[K comparable, V any] func(key K, value V)

// LRU implements a non-thread safe fixed size LRU cache
type LRU[K comparable, V any] struct {
	size      int
	evictList *internal.LruList[K, V]
	items     map[K]*internal.Entry[K, V]
	onEvict   EvictCallback[K, V]
}

func New[K comparable, V any](size int) (*LRU[K, V], error) {
	return NewLRU[K, V](size, nil)
}

// NewLRU constructs an LRU of the given size
func NewLRU[K comparable, V any](size int, onEvict EvictCallback[K, V]) (*LRU[K, V], error) {
	if size <= 0 {
		return nil, errors.New("must provide a positive size")
	}

	c := &LRU[K, V]{
		size:      size,
		evictList: internal.NewList[K, V](),
		items:     make(map[K]*internal.Entry[K, V], size),
		onEvict:   onEvict,
	}
	return c, nil
}

// Purge is used to completely clear the cache.
func (c *LRU[K, V]) Purge() {
	for k, v := range c.items {
		if c.onEvict != nil {
			c.onEvict(k, v.Value)
		}
		delete(c.items, k)
	}
	c.evictList.Init()
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *LRU[K, V]) Add(key K, value V) (evicted bool) {
	// Check for existing item
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		ent.Value = value
		return false
	}

	// Add new item
	ent := c.evictList.PushFront(key, value)
	c.items[key] = ent

	evict := c.evictList.Length() > c.size
	// Verify size not exceeded
	if evict {
		c.removeOldest()
	}
	return evict
}

// AddMany adds multiple values to the cache. Returns the number of evicted items.
func (c *LRU[K, V]) AddMany(keys []K, values []V) (evicted int) {
	// check if keys and values have the same length
	if len(keys) != len(values) {
		return 0
	}

	for i := 0; i < len(keys); i++ {
		key := keys[i]
		value := values[i]

		if ent, ok := c.items[key]; ok {
			c.evictList.MoveToFront(ent)
			ent.Value = value
			continue
		}

		// add new item
		ent := c.evictList.PushFront(key, value)
		c.items[key] = ent

		if c.evictList.Length() > c.size {
			c.removeOldest()
			evicted++
		}
	}

	return evicted
}

// Get looks up a key's value from the cache.
func (c *LRU[K, V]) Get(key K) (value V, ok bool) {
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		return ent.Value, true
	}
	return
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *LRU[K, V]) Contains(key K) (ok bool) {
	_, ok = c.items[key]
	return ok
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *LRU[K, V]) Peek(key K) (value V, ok bool) {
	var ent *internal.Entry[K, V]
	if ent, ok = c.items[key]; ok {
		return ent.Value, true
	}
	return
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *LRU[K, V]) Remove(key K) (present bool) {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return true
	}
	return false
}

func (c *LRU[K, V]) RemoveMany(keys []K) (removed int) {
	for _, key := range keys {
		if ent, ok := c.items[key]; ok {
			c.removeElement(ent)
			removed++
		}
	}
	return removed
}

// RemoveOldest removes the oldest item from the cache.
func (c *LRU[K, V]) RemoveOldest() (key K, value V, ok bool) {
	if ent := c.evictList.Back(); ent != nil {
		c.removeElement(ent)
		return ent.Key, ent.Value, true
	}
	return
}

// GetOldest returns the oldest entry
func (c *LRU[K, V]) GetOldest() (key K, value V, ok bool) {
	if ent := c.evictList.Back(); ent != nil {
		return ent.Key, ent.Value, true
	}
	return
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LRU[K, V]) Keys(reverse bool) []K {
	keys := make([]K, c.evictList.Length())
	i := 0
	if reverse == true {
		for ent := c.evictList.Front(); ent != nil; ent = ent.NextEntry() {
			keys[i] = ent.Key
			i++
		}
	} else {
		for ent := c.evictList.Back(); ent != nil; ent = ent.PrevEntry() {
			keys[i] = ent.Key
			i++
		}
	}
	return keys
}

// Values returns a slice of the values in the cache, from oldest to newest.
func (c *LRU[K, V]) Values(reverse bool) []V {
	values := make([]V, len(c.items))
	i := 0
	if reverse == true {
		for ent := c.evictList.Front(); ent != nil; ent = ent.NextEntry() {
			values[i] = ent.Value
			i++
		}
	} else {
		for ent := c.evictList.Back(); ent != nil; ent = ent.PrevEntry() {
			values[i] = ent.Value
			i++
		}
	}
	return values
}

// Len returns the number of items in the cache.
func (c *LRU[K, V]) Len() int {
	return c.evictList.Length()
}

// Resize changes the cache size.
func (c *LRU[K, V]) Resize(size int) (evicted int, err error) {
	if size <= 0 {
		return c.Len() - size, errors.New("must provide a positive size")
	}
	diff := c.Len() - size
	if diff < 0 {
		diff = 0
	}
	for i := 0; i < diff; i++ {
		c.removeOldest()
	}
	c.size = size
	return diff, nil
}

// removeOldest removes the oldest item from the cache.
func (c *LRU[K, V]) removeOldest() {
	if ent := c.evictList.Back(); ent != nil {
		c.removeElement(ent)
	}
}

// removeElement is used to remove a given list element from the cache
func (c *LRU[K, V]) removeElement(e *internal.Entry[K, V]) {
	c.evictList.Remove(e)
	delete(c.items, e.Key)
	if c.onEvict != nil {
		c.onEvict(e.Key, e.Value)
	}
}

// Keys returns a slice of the keys in the cache.
// The frequently used keys are first in the returned slice.
func (c *TwoQueueCache[K, V]) Keys(reverse bool) []K {
	c.lock.RLock()
	defer c.lock.RUnlock()
	k1 := c.frequent.Keys(reverse)
	k2 := c.recent.Keys(reverse)
	return append(k1, k2...)
}

// Values returns a slice of the values in the cache.
// The frequently used values are first in the returned slice.
func (c *TwoQueueCache[K, V]) Values(reverse bool) []V {
	c.lock.RLock()
	defer c.lock.RUnlock()
	v1 := c.frequent.Values(reverse)
	v2 := c.recent.Values(reverse)
	return append(v1, v2...)
}

// Remove removes the provided key from the cache.
func (c *TwoQueueCache[K, V]) Remove(key K) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.frequent.Remove(key) {
		return
	}
	if c.recent.Remove(key) {
		return
	}
	if c.recentEvict.Remove(key) {
		return
	}
}

// Purge is used to completely clear the cache.
func (c *TwoQueueCache[K, V]) Purge() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.recent.Purge()
	c.frequent.Purge()
	c.recentEvict.Purge()
}

// Contains is used to check if the cache contains a key
// without updating recency or frequency.
func (c *TwoQueueCache[K, V]) Contains(key K) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.frequent.Contains(key) || c.recent.Contains(key)
}

// Peek is used to inspect the cache value of a key
// without updating recency or frequency.
func (c *TwoQueueCache[K, V]) Peek(key K) (value V, ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if val, ok := c.frequent.Peek(key); ok {
		return val, ok
	}
	return c.recent.Peek(key)
}
