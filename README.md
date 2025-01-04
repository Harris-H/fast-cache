# fast-cache

该项目基于[golang-lru](https://github.com/hashicorp/golang-lru)二次开发，是其简化版，并进行了一些修改。



## 1 新增的特性

- 支持缓存由新到旧遍历Key、Value(由reverse参数驱动)
- 对Resize()函数添加错误处理(当size为负数报错)
- 新增AddMany方法，可以一次性添加多个(key,value)对，提高性能。
- 新增RemoveMany方法，一次性删除多个(key,value)对。
- 支持改进的2q

## 2 使用示例

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



## 3 数据结构

### 2Q

simple 2Q算法类似LRU-2，不同点在于2Q将LRU-2算法中的访问历史队列（仅作记录不缓存数据）改为FIFO缓存队列。即，simple 2Q算法有两个缓存队列，一个FIFO队列，一个LRU队列。

<img src=".\assets\2q.png" alt="2q" style="zoom:50%;" />

(1). 新访问的数据插入到FIFO队列；

(2). 如果数据在FIFO队列中一直没有被再次访问，则最终按照FIFO规则淘汰；

(3). 如果数据在FIFO队列中被再次访问，则将数据移到LRU队列头部；

(4). 如果数据在LRU队列再次被访问，则将数据移到LRU队列头部；

(5). LRU队列淘汰末尾的数据。

- FIFO缓存队列在LRU队列缓存之前充当过滤器，任何尝试进入LRU缓存的数据都必须首先通过此传入缓冲区。
- 只有当再次访问一个 Item 时，它才会从FIFO队列提升到 LRU 缓存队列。

而本项目是基于优化的2q：

- 当FIFO队列已满，并且尝试再加入一个新数据到FIFO中时，此时会淘汰FIFO的队头数据，我们不会立即驱逐该项目，而是将其移动到另一个缓冲区中，我们称之为驱逐缓冲区(Evict Buffer)。
  - 该缓冲区将保留已经被淘汰的数据，直到它也已满，如果此时Evict Buffer中有数据被再次访问，则将其从Evice Buffer中删除，并加入到LRU队列中。

> 如果数据遵循整齐且可预测的正态分布，则LRU可能工作正常。
>
> 但现实世界很少如此这样，它充满了长尾场景，例如搜索查询、电子商务推荐，以及少数项目获得大量关注而其余项目仍然是小众项目的任何内容。
>
> 2Q可帮助缓存专注于重要的命中，从而提高性能并免于不必要的麻烦。

## 4 待完善

- 支持lfu
- 2q可获取当前Evict Buffer的数据