package jobs

import (
	"context"
	"fmt"
)

type JobDispatcher struct {
	ch       chan<- JobFunc
}

func NewDefaultJobDispatcher() *JobDispatcher {
	return NewDefaultJobDispatcherContext(context.Background())
}

func NewDefaultJobDispatcherContext(ctx context.Context) *JobDispatcher {
	return NewJobDispatcher(ctx, 10, 10)
}

func NewJobDispatcher(ctx context.Context, count, subCount int) *JobDispatcher {
	ch := make(chan JobFunc, count*100)
	queue := make(chan *jobWorker, count)

	dispatcher := &JobDispatcher{
		ch:       ch,
	}

	for i := 0; i < count; i++ {
		queue <- newJobWorker(ctx, subCount)
		dn := fmt.Sprintf("dispatcher-%d", i)
		go handle(context.WithValue(ctx, "name", dn), ch, queue)
	}

	return dispatcher
}

func handle(ctx context.Context, ch <-chan JobFunc, queue chan *jobWorker) {
	for {
		select {
		case <-ctx.Done():
			return
		case j := <-ch:
			h := <-queue
			for !h.add(j) {
				queue <- h
				h = <-queue
			}
			queue <- h
		}
	}
}

func (d *JobDispatcher) AddJob(job Job) {
	d.ch <- job.Exec
}

func (d *JobDispatcher) AddFunc(f func()) {
	d.ch <- f
}
