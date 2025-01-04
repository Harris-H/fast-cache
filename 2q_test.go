package main

import (
	"crypto/rand"
	"fast-cache/lru"
	"math"
	"math/big"
	"testing"
)

func getRand(tb testing.TB) int64 {
	out, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		tb.Fatal(err)
	}
	return out.Int64()
}

func Test2Q(t *testing.T) {
	l, err := lru.New2Q[int, string](5)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	l.Add(1, "Java")
	l.Add(2, "Go")
	l.Add(3, "Python")
	l.Add(4, "C++")
	l.Add(5, "C")
	keys := l.Keys(false)
	t.Logf("keys: %v", keys)
	keysOrderedByNew := l.Keys(true)
	t.Logf("keysOrderedByNew 1: %v", keysOrderedByNew)
	value, ok := l.Get(3)
	if ok {
		t.Logf("key: %v value: %v", 3, value)
	}
	t.Logf("keysOrderedByNew 2: %v", l.Keys(true))
	l.Add(6, "Rust")
	keysOrderedByNew = l.Keys(true)
	t.Logf("keysOrderedByNew 3: %v", keysOrderedByNew)
	l.Remove(5)
	t.Logf("keysOrderedByNew 4: %v", l.Keys(true))
	/*
	   keys: [1 2 3 4 5]
	   keysOrderedByNew 1: [5 4 3 2 1]
	   key: 3 value: Python
	   keysOrderedByNew 2: [3 5 4 2 1]
	   keysOrderedByNew 3: [3 6 5 4 2]
	   keysOrderedByNew 4: [3 6 4 2]
	*/
}
func Benchmark2Q_Rand(b *testing.B) {
	l, err := lru.New2Q[int64, int64](8192)
	if err != nil {
		b.Fatalf("err: %v", err)
	}

	trace := make([]int64, b.N*2)
	for i := 0; i < b.N*2; i++ {
		trace[i] = getRand(b) % 32768
	}

	b.ResetTimer()

	var hit, miss int
	for i := 0; i < 2*b.N; i++ {
		if i%2 == 0 {
			l.Add(trace[i], trace[i])
		} else {
			if _, ok := l.Get(trace[i]); ok {
				hit++
			} else {
				miss++
			}
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(hit+miss))
}

func Benchmark2Q_Freq(b *testing.B) {
	l, err := lru.New2Q[int64, int64](8192)
	if err != nil {
		b.Fatalf("err: %v", err)
	}

	trace := make([]int64, b.N*2)
	for i := 0; i < b.N*2; i++ {
		if i%2 == 0 {
			trace[i] = getRand(b) % 16384
		} else {
			trace[i] = getRand(b) % 32768
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		l.Add(trace[i], trace[i])
	}
	var hit, miss int
	for i := 0; i < b.N; i++ {
		if _, ok := l.Get(trace[i]); ok {
			hit++
		} else {
			miss++
		}
	}
	b.Logf("hit: %d miss: %d ratio: %f", hit, miss, float64(hit)/float64(hit+miss))
}

func Test2Q_RandomOps(t *testing.T) {
	size := 128
	l, err := lru.New2Q[int64, int64](128)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	n := 200000
	for i := 0; i < n; i++ {
		key := getRand(t) % 512
		r := getRand(t)
		switch r % 3 {
		case 0:
			l.Add(key, key)
		case 1:
			l.Get(key)
		case 2:
			l.Remove(key)
		}

		if l.Len() > size {
			t.Fatalf("bad: recent+freq: %d",
				l.Len())
		}
	}
}

// Test that Peek doesn't update recent-ness
func Test2Q_Peek(t *testing.T) {
	l, err := lru.New2Q[int, int](2)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	l.Add(1, 1)
	l.Add(2, 2)
	if v, ok := l.Peek(1); !ok || v != 1 {
		t.Errorf("1 should be set to 1: %v, %v", v, ok)
	}

	l.Add(3, 3)
	if l.Contains(1) {
		t.Errorf("should not have updated recent-ness of 1")
	}
}
