package lru

import (
	"errors"
	"sync"
)

const (
	// Default2QRecentRatio is the ratio of the 2Q cache dedicated
	// to recently added entries that have only been accessed once.
	Default2QRecentRatio = 0.25

	// Default2QGhostEntries is the default ratio of ghost
	// entries kept to track entries recently evicted
	Default2QGhostEntries = 0.50
)

// TwoQueueCache is a thread-safe fixed size 2Q cache.
// 2Q is an enhancement over the standard LRU cache
// in that it tracks both frequently and recently used
// entries separately. This avoids a burst in access to new
// entries from evicting frequently used entries. It adds some
// additional tracking overhead to the standard LRU cache, and is
// computationally about 2x the cost, and adds some metadata over
// head. The ARCCache is similar, but does not require setting any
// parameters.
type TwoQueueCache[K comparable, V any] struct {
	size        int
	recentSize  int
	recentRatio float64
	ghostRatio  float64

	recent      Cache[K, V]
	frequent    Cache[K, V]
	recentEvict Cache[K, struct{}]
	lock        sync.RWMutex
}

// New2Q creates a new TwoQueueCache using the default
// values for the parameters.
func New2Q[K comparable, V any](size int) (*TwoQueueCache[K, V], error) {
	return New2QParams[K, V](size, Default2QRecentRatio, Default2QGhostEntries)
}

// New2QParams creates a new TwoQueueCache using the provided
// parameter values.
func New2QParams[K comparable, V any](size int, recentRatio, ghostRatio float64) (*TwoQueueCache[K, V], error) {
	if size <= 0 {
		return nil, errors.New("invalid size")
	}
	if recentRatio < 0.0 || recentRatio > 1.0 {
		return nil, errors.New("invalid recent ratio")
	}
	if ghostRatio < 0.0 || ghostRatio > 1.0 {
		return nil, errors.New("invalid ghost ratio")
	}

	// Determine the sub-sizes
	recentSize := int(float64(size) * recentRatio)
	evictSize := int(float64(size) * ghostRatio)

	// Allocate the LRUs
	recent, err := NewLRU[K, V](size, nil)
	if err != nil {
		return nil, err
	}
	frequent, err := NewLRU[K, V](size, nil)
	if err != nil {
		return nil, err
	}
	recentEvict, err := NewLRU[K, struct{}](evictSize, nil)
	if err != nil {
		return nil, err
	}

	// Initialize the cache
	c := &TwoQueueCache[K, V]{
		size:        size,
		recentSize:  recentSize,
		recentRatio: recentRatio,
		ghostRatio:  ghostRatio,
		recent:      recent,
		frequent:    frequent,
		recentEvict: recentEvict,
	}
	return c, nil
}

// Get looks up a key's value from the cache.
func (c *TwoQueueCache[K, V]) Get(key K) (value V, ok bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Check if this is a frequent value
	if val, ok := c.frequent.Get(key); ok {
		return val, ok
	}

	// If the value is contained in recent, then we
	// promote it to frequent
	if val, ok := c.recent.Peek(key); ok {
		c.recent.Remove(key)
		c.frequent.Add(key, val)
		return val, ok
	}

	// No hit
	return
}

// Add adds a value to the cache.
func (c *TwoQueueCache[K, V]) Add(key K, value V) {
	c.lock.Lock()
	defer c.lock.Unlock()

	// Check if the value is frequently used already,
	// and just update the value
	if c.frequent.Contains(key) {
		c.frequent.Add(key, value)
		return
	}

	// Check if the value is recently used, and promote
	// the value into the frequent list
	if c.recent.Contains(key) {
		c.recent.Remove(key)
		c.frequent.Add(key, value)
		return
	}

	// If the value was recently evicted, add it to the
	// frequently used list
	if c.recentEvict.Contains(key) {
		c.ensureSpace(true)
		c.recentEvict.Remove(key)
		c.frequent.Add(key, value)
		return
	}

	// Add to the recently seen list
	c.ensureSpace(false)
	c.recent.Add(key, value)
}

// ensureSpace is used to ensure we have space in the cache
func (c *TwoQueueCache[K, V]) ensureSpace(recentEvict bool) {
	// If we have space, nothing to do
	recentLen := c.recent.Len()
	freqLen := c.frequent.Len()
	if recentLen+freqLen < c.size {
		return
	}

	// If the recent buffer is larger than
	// the target, evict from there
	if recentLen > 0 && (recentLen > c.recentSize || (recentLen == c.recentSize && !recentEvict)) {
		k, _, _ := c.recent.RemoveOldest()
		c.recentEvict.Add(k, struct{}{})
		return
	}

	// Remove from the frequent list otherwise
	c.frequent.RemoveOldest()
}

// Len returns the number of items in the cache.
func (c *TwoQueueCache[K, V]) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.recent.Len() + c.frequent.Len()
}

// Resize changes the cache size.
func (c *TwoQueueCache[K, V]) Resize(size int) (evicted int, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if size <= 0 {
		return c.recent.Len() + c.frequent.Len() - size, errors.New("must provide a positive size")
	}
	// Recalculate the sub-sizes
	recentSize := int(float64(size) * c.recentRatio)
	evictSize := int(float64(size) * c.ghostRatio)
	c.size = size
	c.recentSize = recentSize

	// ensureSpace
	diff := c.recent.Len() + c.frequent.Len() - size
	if diff < 0 {
		diff = 0
	}
	for i := 0; i < diff; i++ {
		c.ensureSpace(true)
	}

	// Reallocate the LRUs
	_, _ = c.recent.Resize(size)
	_, _ = c.frequent.Resize(size)
	_, _ = c.recentEvict.Resize(evictSize)
	return diff, nil
}
