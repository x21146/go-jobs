package jobs

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestJobDispatcher(t *testing.T) {
	count := 10000000
	wg := sync.WaitGroup{}
	wg.Add(count)

	c := make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	dispatcher := NewDefaultJobDispatcherContext(ctx)

	var counter int32 = 0

	// add goroutine
	go func() {
		for i := 0; i < count; i++ {
			dispatcher.AddFunc(func() {
				atomic.AddInt32(&counter, 1)
				wg.Done()
			})
		}
	}()

	// wait group goroutine
	go func() {
		wg.Wait()
		close(c)
	}()

	select {
	case <-time.After(1 * time.Second):
	case <-c:
	}

	cancel()
	t.Log(counter)
}

func BenchmarkJobDispatcher(b *testing.B) {
	wg := sync.WaitGroup{}
	dispatcher := NewDefaultJobDispatcher()

	b.ResetTimer()

	wg.Add(b.N)
	for i := 0; i < b.N; i++ {
		dispatcher.AddFunc(func() {
			wg.Done()
		})
	}

	wg.Wait()
}
