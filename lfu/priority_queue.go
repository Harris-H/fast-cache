package lfu

import (
	"container/heap"
	"time"
)

type PqEntry[K comparable, V any] struct {
	index          int
	Key            K
	Val            V
	referenceCount int
	referencedAt   time.Time
}

//// GetReferenceCount gets reference count from cache value.
//func GetReferenceCount(v any) int {
//	if getter, ok := v.(interface{ GetReferenceCount() int }); ok {
//		return getter.GetReferenceCount()
//	}
//	return 1
//}

func newEntry[K comparable, V any](key K, val V) *PqEntry[K, V] {
	return &PqEntry[K, V]{
		index:          0,
		Key:            key,
		Val:            val,
		referenceCount: 1,
		referencedAt:   time.Now(),
	}
}

func (e *PqEntry[K, V]) referenced() {
	e.referenceCount++
	e.referencedAt = time.Now()
}

type PriorityQueue[K comparable, V any] []*PqEntry[K, V]

func NewPriorityQueue[K comparable, V any](cap int) *PriorityQueue[K, V] {
	queue := make(PriorityQueue[K, V], 0, cap)
	return &queue
}

// see example of priority queue: https://pkg.go.dev/container/heap
var _ heap.Interface = (*PriorityQueue[struct{}, interface{}])(nil)

func (q PriorityQueue[K, V]) Len() int { return len(q) }

func (q PriorityQueue[K, V]) Less(i, j int) bool {
	if q[i].referenceCount == q[j].referenceCount {
		return q[i].referencedAt.Before(q[j].referencedAt)
	}
	return q[i].referenceCount < q[j].referenceCount
}

func (q PriorityQueue[K, V]) Swap(i, j int) {
	if len(q) < 2 {
		return
	}
	q[i], q[j] = q[j], q[i]
	q[i].index = i
	q[j].index = j
}

func (q *PriorityQueue[K, V]) Push(x interface{}) {
	entry := x.(*PqEntry[K, V])
	entry.index = len(*q)
	*q = append(*q, entry)
}

func (q *PriorityQueue[K, V]) Pop() interface{} {
	old := *q
	n := len(old)
	if n == 0 {
		return nil // Return nil if the queue is empty to prevent panic
	}
	entry := old[n-1]
	old[n-1] = nil   // avoid memory leak
	entry.index = -1 // for safety
	new := old[0 : n-1]
	*q = new
	return entry
}

func (q *PriorityQueue[K, V]) update(e *PqEntry[K, V], val V) {
	e.Val = val
	e.referenced()
	heap.Fix(q, e.index)
}
