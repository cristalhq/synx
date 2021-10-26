package synx

import (
	"time"
)

type WorkerPool struct {
	lifetime  time.Duration
	semaphore chan struct{}
	taskQueue chan func()
}

func NewWorkerPool(maxWorkers int, lifetime time.Duration) *WorkerPool {
	return &WorkerPool{
		lifetime:  lifetime,
		semaphore: make(chan struct{}, maxWorkers),
		taskQueue: make(chan func()),
	}
}

func (wp *WorkerPool) Do(task func()) {
	select {
	case wp.taskQueue <- task:
		// submitted, everything is ok

	case wp.semaphore <- struct{}{}:
		go wp.startWorker(task)
	}
}

func (wp *WorkerPool) startWorker(task func()) {
	defer func() { <-wp.semaphore }()

	task() // do the given task

	tick := time.NewTicker(wp.lifetime)
	defer tick.Stop()

	for {
		select {
		case t := <-wp.taskQueue:
			t()
			tick.Reset(wp.lifetime)
		case <-tick.C:
			return
		}
	}
}
