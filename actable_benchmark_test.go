package action_test

import (
	"sync"
	"testing"

	"github.com/neonima/action"
)

func BenchmarkActable_Get(b *testing.B) {
	a := action.Actable[int]{Value: b.N}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a.Get()
		}
	})
}

func BenchmarkActable_Mutext(b *testing.B) {
	a := b.N
	var m sync.Mutex
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			m.Lock()
			a = a
			m.Unlock()
		}
	})
}

func BenchmarkActable_Set(b *testing.B) {
	a := action.Actable[int]{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a.Set(b.N)
		}
	})
}

func BenchmarkActable_SetGet(b *testing.B) {
	a := action.Actable[int]{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			a.Set(a.Get() * b.N)
		}
	})
}
