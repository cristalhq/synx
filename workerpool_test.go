package synx

import (
	"runtime"
	"testing"
	"time"
)

func TestWorkerPool(t *testing.T) {
	const shortDelay = 100 * time.Millisecond
	const workers = 3
	wp := NewWorkerPool(workers, time.Second)

	// submit tasks and grow till `workers` goroutines
	startG := runtime.NumGoroutine()

	for i := 1; i <= workers; i++ {
		wp.Do(func() {
			time.Sleep(shortDelay)
		})

		if n := runtime.NumGoroutine() - startG; n != i {
			t.Fatalf("want %v, got %v", i, n)
		}
	}

	// workers are alive so blocking and waiting to finish

	for i := 1; i <= workers; i++ {
		startG = runtime.NumGoroutine()
		wp.Do(func() {
			time.Sleep(shortDelay)
		})

		if n := runtime.NumGoroutine() - startG; n != 0 {
			t.Fatalf("want %v, got %v", i, n)
		}
	}
}
