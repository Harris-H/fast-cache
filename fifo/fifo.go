package fifo

import (
	"errors"
	"fast-cache/internal"
)

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback[K comparable, V any] func(key K, value V)

type FIFO[K comparable, V any] struct {
	size      int
	evictList *internal.LruList[K, V]
	items     map[K]*internal.Entry[K, V]
	onEvict   EvictCallback[K, V]
}

// NewFIFO constructs an FIFO of the given size
func NewFIFO[K comparable, V any](size int, onEvict EvictCallback[K, V]) (*FIFO[K, V], error) {
	if size <= 0 {
		return nil, errors.New("must provide a positive size")
	}

	c := &FIFO[K, V]{
		size:      size,
		evictList: internal.NewList[K, V](),
		items:     make(map[K]*internal.Entry[K, V], size),
		onEvict:   onEvict,
	}
	return c, nil
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *FIFO[K, V]) Add(key K, value V) (evicted bool) {
	// Check for existing item
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		ent.Value = value
		return false
	}

	// Add new item
	ent := c.evictList.PushBack(key, value)
	c.items[key] = ent

	evict := c.evictList.Length() > c.size
	// Verify size not exceeded
	if evict {
		c.removeFront()
	}
	return evict
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *FIFO[K, V]) Remove(key K) (present bool) {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return true
	}
	return false
}

// removeOldest removes the oldest item from the cache.
func (c *FIFO[K, V]) removeFront() {
	if ent := c.evictList.Front(); ent != nil {
		c.removeElement(ent)
	}
}

// removeElement is used to remove a given list element from the cache
func (c *FIFO[K, V]) removeElement(e *internal.Entry[K, V]) {
	c.evictList.Remove(e)
	delete(c.items, e.Key)
	if c.onEvict != nil {
		c.onEvict(e.Key, e.Value)
	}
}

// Get looks up a key's value from the cache.
func (c *FIFO[K, V]) Get(key K) (value V, ok bool) {
	if ent, ok := c.items[key]; ok {
		return ent.Value, true
	}
	return
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *FIFO[K, V]) Contains(key K) (ok bool) {
	_, ok = c.items[key]
	return ok
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *FIFO[K, V]) Keys(reverse bool) []K {
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
func (c *FIFO[K, V]) Values(reverse bool) []V {
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
func (c *FIFO[K, V]) Len() int {
	return c.evictList.Length()
}

// Resize changes the cache size.
func (c *FIFO[K, V]) Resize(size int) (evicted int, err error) {
	if size <= 0 {
		return c.Len() - size, errors.New("must provide a positive size")
	}
	diff := c.Len() - size
	if diff < 0 {
		diff = 0
	}
	for i := 0; i < diff; i++ {
		c.removeFront()
	}
	c.size = size
	return diff, nil
}
