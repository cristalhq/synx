package synx

import (
	"context"
	"sync"
)

// Signal channel.
type Signal = chan struct{}

// BlockForever the calling goroutine.
func BlockForever() {
	select {}
}

// Async executes fn in a goroutine.
// Returned channel is closed when goroutine completes.
func Async(fn func()) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		fn()
	}()
	return ch
}

// Wait for a function to finish.
func Wait(ctx context.Context, fn func()) error {
	ch := Async(fn)

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
		return nil
	}
}

// Locked call of fn.
func Locked(mu *sync.Mutex, fn func()) {
	mu.Lock()
	defer mu.Unlock()
	fn()
}

// RLocked call of fn.
func RLocked(mu *sync.RWMutex, fn func()) {
	mu.RLock()
	defer mu.RUnlock()
	fn()
}
