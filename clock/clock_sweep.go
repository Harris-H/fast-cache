package clock

import (
	"container/ring"
	"errors"
)

type CSEntry[K comparable, V any] struct {
	Key      K
	Val      V
	refCount int
	useCount int
}

type ClockSweep[K comparable, V any] struct {
	size    int
	items   map[K]*ring.Ring
	hand    *ring.Ring
	head    *ring.Ring
	onEvict EvictCallback[K, V]
}

// NewClockSweep constructs an Clock of the given size
func NewClockSweep[K comparable, V any](size int, onEvict EvictCallback[K, V]) (*ClockSweep[K, V], error) {
	if size <= 0 {
		return nil, errors.New("must provide a positive size")
	}
	r := ring.New(size)
	c := &ClockSweep[K, V]{
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
func (c *ClockSweep[K, V]) Add(key K, val V) {
	if e, ok := c.items[key]; ok {
		entry := e.Value.(*CSEntry[K, V])
		entry.useCount++
		entry.Val = val
		return
	}
	c.evict()
	c.hand.Value = &CSEntry[K, V]{
		Key:      key,
		Val:      val,
		refCount: 1,
		useCount: 0,
	}
	c.items[key] = c.hand
	c.hand = c.hand.Next()
}

// Get looks up a key's value from the cache.
func (c *ClockSweep[K, V]) Get(key K) (value V, ok bool) {
	if ent, ok := c.items[key]; ok {
		entry := ent.Value.(*CSEntry[K, V])
		entry.useCount++
		return entry.Val, true
	}
	return
}

func (c *ClockSweep[K, V]) evict() {
	for c.hand.Value != nil {
		if c.hand.Value.(*CSEntry[K, V]).refCount == 0 {
			if c.hand.Value.(*CSEntry[K, V]).useCount > 0 {
				c.hand.Value.(*CSEntry[K, V]).useCount--
				c.hand = c.hand.Next()
			} else {
				break
			}
		} else {
			// avoid infinite loop
			c.hand.Value.(*CSEntry[K, V]).refCount--
			c.hand = c.hand.Next()
		}
	}
	if c.hand.Value != nil {
		entry := c.hand.Value.(*CSEntry[K, V])
		delete(c.items, entry.Key)
		c.hand.Value = nil
	}
}

// Keys returns the keys of the cache. the order as same as current ring order.
func (c *ClockSweep[K, V]) Keys() []K {
	keys := make([]K, 0, len(c.items))
	r := c.head
	if r.Value == nil {
		return []K{}
	}
	// the first element
	keys = append(keys, r.Value.(*CSEntry[K, V]).Key)

	// iterating
	for p := c.head.Next(); p != r; p = p.Next() {
		if p.Value == nil {
			continue
		}
		e := p.Value.(*CSEntry[K, V])
		keys = append(keys, e.Key)
	}
	return keys
}

// Delete deletes the item with provided key from the cache.
func (c *ClockSweep[K, V]) Delete(key K) {
	if e, ok := c.items[key]; ok {
		delete(c.items, key)
		e.Value = nil
		if c.onEvict != nil {
			c.onEvict(e.Value.(*CSEntry[K, V]).Key, e.Value.(*CSEntry[K, V]).Val)
		}
	}
}

// Len returns the number of items in the cache.
func (c *ClockSweep[K, V]) Len() int {
	return len(c.items)
}
