package main

import (
	"fast-cache/lru"
	"fmt"
)

func main() {
	l, _ := lru.New[int, any](5)
	l.AddMany([]int{1, 2, 3, 4, 5}, []any{1, 2, 3, 4, 5})
	fmt.Println(l.GetOldest())
	l.Get(1)
	fmt.Println(l.GetOldest())
	l.Add(6, 6*5)
	fmt.Println(l.GetOldest())
	_, err := l.Resize(-1)
	if err != nil {
		fmt.Println(err)
		return
	}
	l.Add(1, 1)
}
