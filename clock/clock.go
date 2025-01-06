package clock

import (
	"container/ring"
	"errors"
)

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback[K comparable, V any] func(key K, value V)

type CEntry[K comparable, V any] struct {
	Key      K
	Val      V
	refCount int
}

type Clock[K comparable, V any] struct {
	size    int
	items   map[K]*ring.Ring
	hand    *ring.Ring
	head    *ring.Ring
	onEvict EvictCallback[K, V]
}

// NewClock constructs an Clock of the given size
func NewClock[K comparable, V any](size int, onEvict EvictCallback[K, V]) (*Clock[K, V], error) {
	if size <= 0 {
		return nil, errors.New("must provide a positive size")
	}
	r := ring.New(size)
	c := &Clock[K, V]{
		size:    size,
		hand:    r,
		head:    r,
		items:   make(map[K]*ring.Ring, size),
		onEvict: onEvict,
	}
	return c, nil
}

// Add Set sets any item to the cache. replacing any existing item.
//
// If value satisfies "interface{ GetReferenceCount() int }", the value of
// the GetReferenceCount() method is used to set the initial value of reference count.
func (c *Clock[K, V]) Add(key K, val V) {
	if e, ok := c.items[key]; ok {
		entry := e.Value.(*CEntry[K, V])
		entry.refCount++
		entry.Val = val
		return
	}
	c.evict()
	c.hand.Value = &CEntry[K, V]{
		Key:      key,
		Val:      val,
		refCount: 1,
	}
	c.items[key] = c.hand
	c.hand = c.hand.Next()
}

// Get looks up a key's value from the cache.
func (c *Clock[K, V]) Get(key K) (value V, ok bool) {
	if ent, ok := c.items[key]; ok {
		entry := ent.Value.(*CEntry[K, V])
		entry.refCount++
		return entry.Val, true
	}
	return
}

func (c *Clock[K, V]) evict() {
	for c.hand.Value != nil && c.hand.Value.(*CEntry[K, V]).refCount > 0 {
		c.hand.Value.(*CEntry[K, V]).refCount--
		c.hand = c.hand.Next()
	}
	if c.hand.Value != nil {
		entry := c.hand.Value.(*CEntry[K, V])
		delete(c.items, entry.Key)
		c.hand.Value = nil
	}
}

// Keys returns the keys of the cache. the order as same as current ring order.
func (c *Clock[K, V]) Keys() []K {
	keys := make([]K, 0, len(c.items))
	r := c.head
	if r.Value == nil {
		return []K{}
	}
	// the first element
	keys = append(keys, r.Value.(*CEntry[K, V]).Key)

	// iterating
	for p := c.head.Next(); p != r; p = p.Next() {
		if p.Value == nil {
			continue
		}
		e := p.Value.(*CEntry[K, V])
		keys = append(keys, e.Key)
	}
	return keys
}

// Delete deletes the item with provided key from the cache.
func (c *Clock[K, V]) Delete(key K) {
	if e, ok := c.items[key]; ok {
		delete(c.items, key)
		e.Value = nil
		if c.onEvict != nil {
			c.onEvict(e.Value.(*CEntry[K, V]).Key, e.Value.(*CEntry[K, V]).Val)
		}
	}
}

// Len returns the number of items in the cache.
func (c *Clock[K, V]) Len() int {
	return len(c.items)
}
