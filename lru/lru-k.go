package lru

import (
	"errors"
	"sync"
)

type LRUK[K comparable, V any] struct {
	size       int
	recentSize int
	k          uint8
	recent     Cache[K, V]
	cnt        map[K]uint8
	frequent   Cache[K, V]
	lock       sync.RWMutex
}

func NewLruK[K comparable, V any](size int, k uint8) (*LRUK[K, V], error) {
	return NewLruKParams[K, V](size, Default2QRecentRatio, k)
}
func NewLruKParams[K comparable, V any](size int, recentRatio float64, k uint8) (*LRUK[K, V], error) {
	if size <= 0 || k <= 0 {
		return nil, errors.New("invalid size or k")
	}
	if recentRatio < 0.0 || recentRatio > 1.0 {
		return nil, errors.New("invalid recent ratio")
	}
	recent, err := NewLRU[K, V](size, nil)
	if err != nil {
		return nil, err
	}
	frequent, err := NewLRU[K, V](size, nil)
	if err != nil {
		return nil, err
	}
	// Determine the sub-sizes
	recentSize := int(float64(size) * recentRatio)
	c := &LRUK[K, V]{
		size:       size,
		recentSize: recentSize,
		k:          k,
		cnt:        make(map[K]uint8, size),
		recent:     recent,
		frequent:   frequent,
	}
	return c, nil
}
func (c *LRUK[K, V]) AddFreq(key K, value V) {
	if c.cnt[key] >= c.k {
		c.recent.Remove(key)
		c.frequent.Add(key, value)
		delete(c.cnt, key)
	} else {
		c.recent.MoveToFront(key)
	}
}
func (c *LRUK[K, V]) Get(key K) (value V, ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if value, ok = c.frequent.Get(key); ok {
		return value, ok
	}
	if value, ok = c.recent.Peek(key); ok {
		c.cnt[key]++
		c.AddFreq(key, value)
		return value, ok
	}
	return
}

func (c *LRUK[K, V]) Add(key K, value V) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if _, ok := c.frequent.Get(key); ok {
		c.frequent.Add(key, value)
		return
	}
	if c.recent.Contains(key) {
		c.recent.MoveToFront(key)
	} else {
		c.recent.Add(key, value)
	}
	c.cnt[key]++
	c.AddFreq(key, value)
}

// Len returns the number of items in the cache.
func (c *LRUK[K, V]) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.recent.Len() + c.frequent.Len()
}
func (c *LRUK[K, V]) ensureSpace() {
	// If we have space, nothing to do
	recentLen := c.recent.Len()
	freqLen := c.frequent.Len()
	if recentLen+freqLen < c.size {
		return
	}

	// If the recent buffer is larger than
	// the target, evict from there
	if recentLen > 0 && (recentLen > c.recentSize || (recentLen == c.recentSize)) {
		_, _, _ = c.recent.RemoveOldest()
		return
	}
	// Remove from the frequent list otherwise
	c.frequent.RemoveOldest()
}

// Resize changes the cache size.
func (c *LRUK[K, V]) Resize(size int) (evicted int, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if size <= 0 {
		return c.recent.Len() + c.frequent.Len() - size, errors.New("must provide a positive size")
	}
	// Recalculate the sub-sizes
	c.size = size
	// ensureSpace
	diff := c.recent.Len() + c.frequent.Len() - size
	if diff < 0 {
		diff = 0
	}
	for i := 0; i < diff; i++ {
		c.ensureSpace()
	}

	// Reallocate the LRUs
	_, _ = c.recent.Resize(size)
	_, _ = c.frequent.Resize(size)
	return diff, nil
}

// Keys returns a slice of the keys in the cache.
// The frequently used keys are first in the returned slice.
func (c *LRUK[K, V]) Keys(reverse bool) []K {
	c.lock.RLock()
	defer c.lock.RUnlock()
	k1 := c.frequent.Keys(reverse)
	k2 := c.recent.Keys(reverse)
	return append(k1, k2...)
}

// Values returns a slice of the values in the cache.
// The frequently used values are first in the returned slice.
func (c *LRUK[K, V]) Values(reverse bool) []V {
	c.lock.RLock()
	defer c.lock.RUnlock()
	v1 := c.frequent.Values(reverse)
	v2 := c.recent.Values(reverse)
	return append(v1, v2...)
}

// Remove removes the provided key from the cache.
func (c *LRUK[K, V]) Remove(key K) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.frequent.Remove(key) {
		return
	}
	if c.recent.Remove(key) {
		return
	}
}

// Purge is used to completely clear the cache.
func (c *LRUK[K, V]) Purge() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.recent.Purge()
	c.frequent.Purge()
}

// Contains is used to check if the cache contains a key
// without updating recency or frequency.
func (c *LRUK[K, V]) Contains(key K) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.frequent.Contains(key) || c.recent.Contains(key)
}

// Peek is used to inspect the cache value of a key
// without updating recency or frequency.
func (c *LRUK[K, V]) Peek(key K) (value V, ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	if val, ok := c.frequent.Peek(key); ok {
		return val, ok
	}
	return c.recent.Peek(key)
}
