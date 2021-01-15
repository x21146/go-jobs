package jobs

import (
	"context"
)

type jobWorker struct {
	ch chan JobFunc
}

func newJobWorker(ctx context.Context, count int) *jobWorker {
	handler := &jobWorker{
		ch: make(chan JobFunc, count*100),
	}

	for i := 0; i < count; i++ {
		go func(ch chan JobFunc) {
			for {
				select {
				case <-ctx.Done():
					return
				case j := <-ch:
					j()
					//case <-time.After(1 * time.Second):
					//	continue
				}
			}
		}(handler.ch)
	}

	return handler
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
