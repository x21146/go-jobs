package jobs

import (
	"context"
	"fmt"
	"sync/atomic"
)

type jobWorker struct {
	ch chan<- JobFunc
}

var id int64 = 0

func newJobWorker(ctx context.Context, count int) *jobWorker {
	n := atomic.AddInt64(&id, 1)
	name := fmt.Sprintf("worker-%d", n)
	ct := context.WithValue(ctx, "name", name)
	ch := make(chan JobFunc, count*100)

	handler := &jobWorker{
		ch: ch,
	}

	for i := 0; i < count; i++ {
		wn := fmt.Sprintf("worker-%d-%d", n, i)
		c := context.WithValue(ct, "name", wn)
		go doWork(c, ch)
	}

	return handler
}

func doWork(ctx context.Context, ch <-chan JobFunc) {
	for {
		select {
		case <-ctx.Done():
			return
		case j := <-ch:
			j()
		}
	}
}

func (h *jobWorker) size() int {
	return len(h.ch)
}

func (h *jobWorker) isFull() bool {
	return len(h.ch) == cap(h.ch)
}

func (h *jobWorker) add(job JobFunc) bool {
	select {
	case h.ch <- job:
		return true
	default:
		return false
	}
}
