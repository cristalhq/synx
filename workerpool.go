package synx

import "time"

var taskQueue = make(chan func())

const workerLifetime = 10 * time.Second

// RunAsync a given task. Task will be executed by a worker pool.
// If there is no free worker - new worker will be created.
// Worker without tasks will be finished after workerLifetime.
func RunAsync(task func()) {
	select {
	case taskQueue <- task:
		// submited, everything is ok

	default:
		go func() {
			// do the given task
			task()

			tick := time.NewTicker(workerLifetime)
			defer tick.Stop()

			for {
				select {
				case t := <-taskQueue:
					t()
					tick.Reset(workerLifetime)
				case <-tick.C:
					return
				}
			}
		}()
	}
}
