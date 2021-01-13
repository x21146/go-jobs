package jobs

import (
	"context"
)

type jobDispatcher struct {
	ch       chan JobFunc
	handlers []*jobWorker
	queue    chan *jobWorker
}

func NewDefaultJobDispatcher() *jobDispatcher {
	return NewDefaultJobDispatcherContext(context.Background())
}

func NewDefaultJobDispatcherContext(ctx context.Context) *jobDispatcher {
	return NewJobDispatcher(ctx, 10, 10)
}

func NewJobDispatcher(ctx context.Context, count, subCount int) *jobDispatcher {
	dispatcher := &jobDispatcher{
		ch:       make(chan JobFunc, count*100),
		handlers: make([]*jobWorker, count),
		queue:    make(chan *jobWorker, count),
	}

	for i := range dispatcher.handlers {
		dispatcher.handlers[i] = newJobWorker(ctx, subCount)
		dispatcher.queue <- dispatcher.handlers[i]
	}

	for i := 0; i < count; i++ {
		go dispatcher.handle(ctx)
	}

	return dispatcher
}

func (d *jobDispatcher) handle(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case j := <-d.ch:
			h := <-d.queue
			for !h.add(j) {
				d.queue <- h
				h = <-d.queue
			}
			d.queue <- h
		}
	}
}

func (d *jobDispatcher) AddJob(job Job) {
	d.ch <- job.Exec
}

func (d *jobDispatcher) AddFunc(f func()) {
	d.ch <- f
}
