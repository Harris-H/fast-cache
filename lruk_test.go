package main

import (
	"fast-cache/lru"
	"fmt"
	"testing"
)

func TestLRUK(t *testing.T) {
	l, _ := lru.NewLruK[int, string](5, 2)
	l.Add(1, "Java")
	l.Add(2, "Go")
	l.Add(3, "Python")
	l.Add(4, "C++")
	l.Add(5, "C")
	keys := l.Keys(false)
	fmt.Println("keys: ", keys)
	fmt.Println("keysOrderedByNew 1: ", l.Keys(true))
	value, ok := l.Get(3)
	if ok {
		fmt.Println("key: ", 3, " value: ", value)
	}
	fmt.Println("keysOrderedByNew 2: ", l.Keys(true))
	fmt.Println("Add (6,Rust): ")
	l.Add(6, "Rust")
	fmt.Println("keysOrderedByNew 3: ", l.Keys(true))
	fmt.Println("Get key: 6")
	fmt.Println(l.Get(6))
	fmt.Println("keysOrderedByNew 4: ", l.Keys(true))
	fmt.Println("Remove key: 5")
	l.Remove(5)
	fmt.Println("keysOrderedByNew 5: ", l.Keys(true))
	fmt.Println(l.Get(4))
	fmt.Println("keysOrderedByNew 6: ", l.Keys(true))

	fmt.Println(l.Values(true))
}
