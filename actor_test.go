package phony

import (
	"testing"
	"unsafe"
)

func TestInboxSize(t *testing.T) {
	var a Inbox
	var q queueElem
	t.Logf("Inbox size: %d, message size: %d", unsafe.Sizeof(a), unsafe.Sizeof(q))
}

func TestBlock(t *testing.T) {
	var a Inbox
	var results []int
	for idx := 0; idx < 1024; idx++ {
		n := idx // Because idx gets mutated in place
		Block(&a, func() {
			results = append(results, n)
		})
	}
	for idx, n := range results {
		if n != idx {
			t.Errorf("value %d != index %d", n, idx)
		}
	}
}

func TestAct(t *testing.T) {
	var a Inbox
	var results []int
	Block(&a, func() {
		for idx := 0; idx < 1024; idx++ {
			n := idx // Because idx gets mutated in place
			a.Act(&a, func() {
				results = append(results, n)
			})
		}
	})
	Block(&a, func() {})
	for idx, n := range results {
		if n != idx {
			t.Errorf("value %d != index %d", n, idx)
		}
	}
}

func BenchmarkBlock(b *testing.B) {
	var a Inbox
	for i := 0; i < b.N; i++ {
		Block(&a, func() {})
	}
}

func BenchmarkAct(b *testing.B) {
	var a Inbox
	done := make(chan struct{})
	idx := 0
	var f func()
	f = func() {
		if idx < b.N {
			idx++
			a.Act(&a, f)
		} else {
			close(done)
		}
	}
	a.Act(nil, f)
	<-done
}

func BenchmarkActFromNil(b *testing.B) {
	var a Inbox
	done := make(chan struct{})
	idx := 0
	var f func()
	f = func() {
		if idx < b.N {
			idx++
			a.Act(nil, f)
		} else {
			close(done)
		}
	}
	a.Act(nil, f)
	<-done
}

func BenchmarkChannel(b *testing.B) {
	done := make(chan struct{})
	ch := make(chan func())
	go func() {
		for f := range ch {
			ch <- f
		}
		close(done)
	}()
	f := func() {}
	for i := 0; i < b.N; i++ {
		ch <- f
		g := <-ch
		g()
	}
	close(ch)
	<-done
}

func BenchmarkBufferedChannel(b *testing.B) {
	ch := make(chan func(), 1)
	f := func() {}
	for i := 0; i < b.N; i++ {
		ch <- f
		g := <-ch
		g()
	}
}
