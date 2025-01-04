package main

import (
	"fast-cache/lru"
	"fmt"
	"testing"
)

func TestLRU(t *testing.T) {
	l, _ := lru.New[int, string](5)
	l.AddMany([]int{1, 2, 3, 4, 5}, []string{"Java", "Go", "Python", "C++", "C"})
	keys := l.Keys(false)
	fmt.Println("keys: ", keys)
	keysOrderedByNew := l.Keys(true)
	fmt.Println("keysOrderedByNew 1: ", keysOrderedByNew)
	value, ok := l.Get(3)
	if ok {
		fmt.Println("key: ", 3, " value: ", value)
	}
	fmt.Println("keysOrderedByNew 2: ", keysOrderedByNew)
	fmt.Println("Add (6,Rust): ")
	l.Add(6, "Rust")
	keysOrderedByNew = l.Keys(true)
	fmt.Println("keysOrderedByNew 3: ", keysOrderedByNew)
	fmt.Println("Remove keys: 2, 4, 3")
	l.RemoveMany([]int{2, 4, 3})
	fmt.Println("keysOrderedByNew 4: ", l.Keys(true))
	/*
		keys:  [1 2 3 4 5]
		keysOrderedByNew 1:  [5 4 3 2 1]
		key:  3  value:  Python
		keysOrderedByNew 2:  [5 4 3 2 1]
		Add (6,Rust):
		keysOrderedByNew 3:  [6 3 5 4 2]
	*/
}
