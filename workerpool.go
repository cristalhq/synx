package synx

import (
	"sync/atomic"
	"time"
)

type WorkerPool struct {
	activeWorkers int64
	maxWorkers    int64
	lifetime      time.Duration
	taskQueue     chan func()
}

func NewWorkerPool(maxWorkers int, lifetime time.Duration) *WorkerPool {
	return &WorkerPool{
		maxWorkers: int64(maxWorkers),
		lifetime:   lifetime,
		taskQueue:  make(chan func()),
	}
}

func (wp *WorkerPool) Do(task func()) {
	select {
	case wp.taskQueue <- task:
		// submitted, everything is ok

	default:
		if wp.maxWorkers <= atomic.LoadInt64(&wp.activeWorkers) {
			wp.taskQueue <- task
			return
		}

		go func() {
			atomic.AddInt64(&wp.activeWorkers, 1)
			defer atomic.AddInt64(&wp.activeWorkers, -1)

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
		}()
	}
}
