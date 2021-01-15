package jobs

import (
	"context"
)

type JobDispatcher struct {
	ch       chan JobFunc
	handlers []*jobWorker
	queue    chan *jobWorker
}

func NewDefaultJobDispatcher() *JobDispatcher {
	return NewDefaultJobDispatcherContext(context.Background())
}

func NewDefaultJobDispatcherContext(ctx context.Context) *JobDispatcher {
	return NewJobDispatcher(ctx, 10, 10)
}

func NewJobDispatcher(ctx context.Context, count, subCount int) *JobDispatcher {
	dispatcher := &JobDispatcher{
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

func (d *JobDispatcher) handle(ctx context.Context) {
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

func (d *JobDispatcher) AddJob(job Job) {
	d.ch <- job.Exec
}

func (d *JobDispatcher) AddFunc(f func()) {
	d.ch <- f
}
