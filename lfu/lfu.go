package lfu

import (
	"container/heap"
	"errors"
)

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback[K comparable, V any] func(key K, value V)

type LFU[K comparable, V any] struct {
	size      int
	evictList *PriorityQueue[K, V]
	items     map[K]*PqEntry[K, V]
	onEvict   EvictCallback[K, V]
}

// NewLFU NewLRU constructs an LRU of the given size
func NewLFU[K comparable, V any](size int, onEvict EvictCallback[K, V]) (*LFU[K, V], error) {
	if size <= 0 {
		return nil, errors.New("must provide a positive size")
	}

	c := &LFU[K, V]{
		size:      size,
		evictList: NewPriorityQueue[K, V](size),
		items:     make(map[K]*PqEntry[K, V], size),
		onEvict:   onEvict,
	}
	return c, nil
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *LFU[K, V]) Add(key K, value V) (evicted bool) {
	// Check for existing item
	if ent, ok := c.items[key]; ok {
		c.evictList.update(ent, value)
		return false
	}
	evict := c.evictList.Len() == c.size
	if evict {
		c.removeElement()
	}

	e := newEntry(key, value)
	heap.Push(c.evictList, e)
	c.items[key] = e

	return evict
}

// removeElement is used to remove a given list element from the cache
func (c *LFU[K, V]) removeElement() {
	ent := heap.Pop(c.evictList)
	if ent != nil {
		delete(c.items, ent.(*PqEntry[K, V]).Key)
		if c.onEvict != nil {
			c.onEvict(ent.(*PqEntry[K, V]).Key, ent.(*PqEntry[K, V]).Val)
		}
	}
}

// Get looks up a key's value from the cache.
func (c *LFU[K, V]) Get(key K) (value V, ok bool) {
	if e, ok := c.items[key]; ok {
		e.referenced()
		heap.Fix(c.evictList, e.index)
		return e.Val, true
	}
	return
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *LFU[K, V]) Contains(key K) (ok bool) {
	_, ok = c.items[key]
	return ok
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *LFU[K, V]) Remove(key K) (present bool) {
	if ent, ok := c.items[key]; ok {
		heap.Remove(c.evictList, ent.index)
		delete(c.items, key)
		return true
	}
	return false
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LFU[K, V]) Keys(reverse bool) []K {
	keys := make([]K, len(c.items))
	if reverse == true {
		for index, ent := range *c.evictList {
			keys[index] = ent.Key
		}
	} else {
		for i := c.evictList.Len() - 1; i >= 0; i-- {
			keys[i] = (*c.evictList)[i].Key
		}
	}
	return keys
}

// Values returns a slice of the values in the cache, from oldest to newest.
func (c *LFU[K, V]) Values(reverse bool) []V {
	values := make([]V, len(c.items))
	if reverse == true {
		for index, ent := range *c.evictList {
			values[index] = ent.Val
		}
	} else {
		for i := c.evictList.Len() - 1; i >= 0; i-- {
			values[i] = (*c.evictList)[i].Val
		}
	}
	return values
}

// Len returns the number of items in the cache.
func (c *LFU[K, V]) Len() int {
	return c.evictList.Len()
}
