# fast-cache

该项目基于[golang-lru](https://github.com/hashicorp/golang-lru)二次开发，是其简化版，并进行了一些修改。



## 新增的特性

- 支持缓存由新到旧遍历Key、Value(由reverse参数驱动)
- 对Resize()函数添加错误处理(当size为负数报错)
- 新增AddMany方法，可以一次性添加多个(key,value)对，提高性能。



## 待完善

- 支持RemoveMany()方法
- 。。。