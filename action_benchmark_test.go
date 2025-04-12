package action_test

import (
	. "github.com/neonima/action"
	"github.com/stretchr/testify/require"
	"math/rand"
	"runtime"
	"sort"
	"sync"
	"testing"
	"time"
)

func BenchmarkGetSet(b *testing.B) {
	r := New()
	require.NoError(b, r.Start(b.Context()))
	m := make(map[time.Time]any)
	b.SetParallelism(runtime.NumCPU() * 10000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			now := time.Now()
			Act(r, func() {
				m[now] = now
			})
			res := ActGet(r, func() any {
				return m[now]
			})
			require.Equal(b, now, res)
		}
	})
}

func BenchmarkMutexGetSet(b *testing.B) {
	m := make(map[time.Time]any)
	var mut sync.RWMutex
	b.SetParallelism(runtime.NumCPU() * 10000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			now := time.Now()
			mut.Lock()
			m[now] = now
			mut.Unlock()
			mut.RLock()
			res := m[now]
			mut.RUnlock()
			require.Equal(b, now, res)
		}
	})
}

type ComplexData struct {
	nums []int
}

func (d *ComplexData) Insert(val int) {
	d.nums = append(d.nums, val)
	// Sorting the slice every time increases complexity.
	sort.Ints(d.nums)
}

func (d *ComplexData) GetMedian() int {
	n := len(d.nums)
	if n == 0 {
		return 0
	}
	return d.nums[n/2]
}

func BenchmarkComplexGetSet(b *testing.B) {
	r := New()
	require.NoError(b, r.Start(b.Context()))
	c := &ComplexData{}
	b.SetParallelism(runtime.NumCPU() * 10000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			Act(r, func() {
				c.Insert(rand.Intn(1000))
			})
			med := ActGet(r, func() any {
				return c.GetMedian()
			})
			require.NotEqual(b, 0, med)
		}
	})
}

func BenchmarkComplexMutexGetSet(b *testing.B) {
	c := &ComplexData{}
	var mut sync.RWMutex
	b.SetParallelism(runtime.NumCPU() * 10000)
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mut.Lock()
			c.Insert(rand.Intn(1000))
			mut.Unlock()
			mut.RLock()
			med := c.GetMedian()
			mut.RUnlock()
			require.NotEqual(b, 0, med)
		}
	})
}
