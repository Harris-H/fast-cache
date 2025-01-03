# fast-cache

该项目基于[golang-lru](https://github.com/hashicorp/golang-lru)二次开发，是其简化版，并进行了一些修改。



## 新增的特性

- 支持缓存由新到旧遍历Key、Value(由reverse参数驱动)
- 对Resize()函数添加错误处理(当size为负数报错)
- 新增AddMany方法，可以一次性添加多个(key,value)对，提高性能。



## 使用示例

```go
package main

import (
	"fast-cache/lru"
	"fmt"
)

func main() {
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
	/*
		keys:  [1 2 3 4 5]
		keysOrderedByNew 1:  [5 4 3 2 1]
		key:  3  value:  Python
		keysOrderedByNew 2:  [5 4 3 2 1]
		Add (6,Rust):
		keysOrderedByNew 3:  [6 3 5 4 2]
	*/
}

```



## 待完善

- 支持RemoveMany()方法
- 。。。