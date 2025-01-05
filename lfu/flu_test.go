package lfu

import (
	"fmt"
	"testing"
)

func TestSet(t *testing.T) {
	// set capacity is 1
	cache, err := NewLFU[string, int](1, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	cache.Add("foo", 1)
	if got := cache.Len(); got != 1 {
		t.Fatalf("invalid length: %d", got)
	}
	if got, ok := cache.Get("foo"); got != 1 || !ok {
		t.Fatalf("invalid value got %d, cachehit %v", got, ok)
	}

	// if over the cap
	cache.Add("bar", 2)
	if got := cache.Len(); got != 1 {
		t.Fatalf("invalid length: %d", got)
	}
	bar, ok := cache.Get("bar")
	if bar != 2 || !ok {
		t.Fatalf("invalid value bar %d, cachehit %v", bar, ok)
	}

	// checks deleted oldest
	if _, ok := cache.Get("foo"); ok {
		t.Fatalf("invalid delete oldest value foo %v", ok)
	}

	// valid: if over the cap but same key
	cache.Add("bar", 100)
	if got := cache.Len(); got != 1 {
		t.Fatalf("invalid length: %d", got)
	}
	bar, ok = cache.Get("bar")
	if bar != 100 || !ok {
		t.Fatalf("invalid replacing value bar %d, cachehit %v", bar, ok)
	}

	//t.Run("with initilal reference count", func(t *testing.T) {
	//	cache, err := NewLFU[string, *tmp](2, nil)
	//	if err != nil {
	//		t.Fatalf("err: %v", err)
	//	}
	//	cache.Add("foo", &tmp{i: 10}) // the highest reference count
	//	cache.Add("foo2", &tmp{i: 2}) // expected eviction
	//	if got := cache.Len(); got != 2 {
	//		t.Fatalf("invalid length: %d", got)
	//	}
	//
	//	cache.Add("foo3", &tmp{i: 3})
	//
	//	// checks deleted the lowest reference count
	//	if _, ok := cache.Get("foo2"); ok {
	//		t.Fatalf("invalid delete oldest value foo2 %v", ok)
	//	}
	//	if _, ok := cache.Get("foo"); !ok {
	//		t.Fatalf("invalid value foo is not found")
	//	}
	//})
}

func TestDelete(t *testing.T) {
	l, err := NewLFU[string, int](2, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	l.Add("foo", 1)
	if got := l.Len(); got != 1 {
		t.Fatalf("invalid length: %d", got)
	}

	l.Remove("foo2")
	if got := l.Len(); got != 1 {
		t.Fatalf("invalid length after deleted does not exist key: %d", got)
	}

	l.Remove("foo")
	if got := l.Len(); got != 0 {
		t.Fatalf("invalid length after deleted: %d", got)
	}
	if _, ok := l.Get("foo"); ok {
		t.Fatalf("invalid get after deleted %v", ok)
	}
}

// check don't panic
func TestIssue33(t *testing.T) {
	cache, err := NewLFU[string, int](2, func(key string, value int) {
		fmt.Println("evicted key:", key, "value:", value)
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	cache.Add("foo", 1)
	cache.Add("foo2", 2)
	cache.Add("foo3", 3)
	cache.Add("foo2", 4)
	fmt.Println(cache.Keys(false))
	fmt.Println(cache.Values(false))
	cache.Remove("foo")
	cache.Remove("foo2")
	cache.Remove("foo3")
	fmt.Println(cache.Keys(false))
}
