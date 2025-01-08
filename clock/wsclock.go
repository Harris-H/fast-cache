package clock

import (
	"container/ring"
	"errors"
	"time"
)

type WSEntry[K comparable, V any] struct {
	Key      K
	Val      V
	refCount int
	age      time.Time
}

type WSClock[K comparable, V any] struct {
	size    int
	items   map[K]*ring.Ring
	limit   time.Duration
	hand    *ring.Ring
	head    *ring.Ring
	onEvict EvictCallback[K, V]
}

// NewWSClock constructs an Clock of the given size
func NewWSClock[K comparable, V any](size int, onEvict EvictCallback[K, V]) (*WSClock[K, V], error) {
	if size <= 0 {
		return nil, errors.New("must provide a positive size")
	}
	r := ring.New(size)
	c := &WSClock[K, V]{
		size:    size,
		hand:    r,
		head:    r,
		items:   make(map[K]*ring.Ring, size),
		limit:   5 * time.Second,
		onEvict: onEvict,
	}
	return c, nil
}

// Add Set sets any item to the cache. replacing any existing item.
//
// If value satisfies "interface{ GetReferenceCount() int }", the value of
// the GetReferenceCount() method is used to set the initial value of reference count.
func (c *WSClock[K, V]) Add(key K, val V) {
	if e, ok := c.items[key]; ok {
		entry := e.Value.(*WSEntry[K, V])
		entry.refCount = 1
		entry.Val = val
		return
	}
	c.evict()
	c.hand.Value = &WSEntry[K, V]{
		Key:      key,
		Val:      val,
		refCount: 1,
		age:      time.Now(),
	}
	c.items[key] = c.hand
	c.hand = c.hand.Next()
}

// Get looks up a key's value from the cache.
func (c *WSClock[K, V]) Get(key K) (value V, ok bool) {
	if ent, ok := c.items[key]; ok {
		entry := ent.Value.(*WSEntry[K, V])
		entry.age = time.Now()
		return entry.Val, true
	}
	return
}

func (c *WSClock[K, V]) evict() {
	r := c.hand
	flag := true
	for c.hand.Value != nil && flag {
		if c.hand.Value.(*WSEntry[K, V]).refCount == 0 {
			if c.hand.Value.(*WSEntry[K, V]).age.Add(c.limit).Before(time.Now()) {
				break
			}
		} else {
			// avoid infinite loop
			c.hand.Value.(*WSEntry[K, V]).refCount = 0
			c.hand = c.hand.Next()
		}
		if c.hand == r {
			flag = false
		}
	}
	if c.hand.Value != nil {
		entry := c.hand.Value.(*WSEntry[K, V])
		delete(c.items, entry.Key)
		c.hand.Value = nil
	}

}

// Keys returns the keys of the cache. the order as same as current ring order.
func (c *WSClock[K, V]) Keys() []K {
	keys := make([]K, 0, len(c.items))
	r := c.head
	if r.Value == nil {
		return []K{}
	}
	// the first element
	keys = append(keys, r.Value.(*WSEntry[K, V]).Key)

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
func (c *WSClock[K, V]) Delete(key K) {
	if e, ok := c.items[key]; ok {
		delete(c.items, key)
		e.Value = nil
		if c.onEvict != nil {
			c.onEvict(e.Value.(*WSEntry[K, V]).Key, e.Value.(*WSEntry[K, V]).Val)
		}
	}
}

// Len returns the number of items in the cache.
func (c *WSClock[K, V]) Len() int {
	return len(c.items)
}
