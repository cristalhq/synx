package synx

import (
	"context"
	"sync"
)

// Signal channel.
type Signal = chan struct{}

var closedChan = func() <-chan struct{} {
	ch := make(chan struct{})
	close(ch)
	return ch
}()

// ClosedChan returns a closed struct{} channel.
func ClosedChan() <-chan struct{} {
	return closedChan
}

// BlockForever the calling goroutine.
func BlockForever() {
	select {}
}

// LockInScope should be used with the defer.
//	defer synx.LockInScope(mu)()
//
func LockInScope(lock *sync.Mutex) (unlock func()) {
	lock.Lock()
	return lock.Unlock
}

// RWLockInScope should be used with the defer.
//	defer synx.RLockInScope(mu)()
//
func RWLockInScope(lock *sync.RWMutex) (unlock func()) {
	lock.Lock()
	return lock.Unlock
}

// RLockInScope should be used with the defer.
//	defer synx.RLockInScope(mu)()
//
func RLockInScope(lock *sync.RWMutex) (unlock func()) {
	lock.RLock()
	return lock.RUnlock
}

// Send to the channel with a context.
func Send(ctx context.Context, ch chan<- interface{}, value interface{}) error {
	select {
	case ch <- value:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Recv from the channel with a context.
func Recv(ctx context.Context, ch <-chan interface{}) (value interface{}, stillOpen bool, err error) {
	select {
	case value, stillOpen = <-ch:
		return value, stillOpen, nil
	case <-ctx.Done():
		return nil, false, ctx.Err()
	}
}

// Wait for a sync.WaitGroup with context.
func Wait(ctx context.Context, wg *sync.WaitGroup) error {
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		wg.Wait()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-ch:
		return nil
	}
}

// With given lock call f.
func With(lock sync.Locker, f func()) {
	lock.Lock()
	defer lock.Unlock()
	f()
}

// WithRead lock do f.
func WithRead(mu *sync.RWMutex, f func()) {
	mu.RLock()
	defer mu.RUnlock()
	f()
}
