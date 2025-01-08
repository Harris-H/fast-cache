package clock

import (
	"fmt"
	"testing"
)

func TestSet(t *testing.T) {
	// set size is 1
	//cache, err := NewClock[string, int](1, nil)
	cache, err := NewClockSweep[string, int](1, nil)
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
		t.Fatalf("invalid eviction the oldest value for foo %v", ok)
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
}
func TestDelete(t *testing.T) {
	//cache, err := NewClock[string, int](1, nil)
	cache, err := NewClockSweep[string, int](1, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	cache.Add("foo", 1)
	if got := cache.Len(); got != 1 {
		t.Fatalf("invalid length: %d", got)
	}

	cache.Delete("foo2")
	if got := cache.Len(); got != 1 {
		t.Fatalf("invalid length after deleted does not exist key: %d", got)
	}

	cache.Delete("foo")
	if got := cache.Len(); got != 0 {
		t.Fatalf("invalid length after deleted: %d", got)
	}
	if _, ok := cache.Get("foo"); ok {
		t.Fatalf("invalid get after deleted %v", ok)
	}
}

func TestExampleNewCache(t *testing.T) {
	//c, err := NewClock[string, int](128, nil)
	c, err := NewClockSweep[string, int](1, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	c.Add("a", 1)
	c.Add("b", 2)
	av, aok := c.Get("a")
	bv, bok := c.Get("b")
	cv, cok := c.Get("c")
	fmt.Println(av, aok)
	fmt.Println(bv, bok)
	fmt.Println(cv, cok)
	c.Delete("a")
	_, aok2 := c.Get("a")
	if !aok2 {
		fmt.Println("key 'a' has been deleted")
	}
	// update
	c.Add("b", 3)
	newbv, _ := c.Get("b")
	fmt.Println(newbv)
	// Output:
	// 1 true
	// 2 true
	// 0 false
	// key 'a' has been deleted
	// 3
}

func TestWSClock_Add(t *testing.T) {
	//c, err := NewWSClock[string, int](128, nil)
	c, err := NewWSClock[string, int](1, nil)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	c.Add("a", 1)
	c.Add("b", 2)
	av, aok := c.Get("a")
	bv, bok := c.Get("b")
	cv, cok := c.Get("c")
	fmt.Println(av, aok)
	fmt.Println(bv, bok)
	fmt.Println(cv, cok)
	c.Delete("a")
	_, aok2 := c.Get("a")
	if !aok2 {
		fmt.Println("key 'a' has been deleted")
	}
	// update
	c.Add("b", 3)
	newbv, _ := c.Get("b")
	fmt.Println(newbv)
	// Output:
	// 1 true
	// 2 true
	// 0 false
	// key 'a' has been deleted
	// 3
}
