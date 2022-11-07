package synx

import (
	"context"
	"sync"
)

// ContextGroup is simple wrapper around sync.WaitGroup.
type ContextGroup struct {
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	errOnce sync.Once
	err     error
}

// NewContextGroup returns new ContextGroup.
func NewContextGroup(parent context.Context) *ContextGroup {
	ctx, cancel := context.WithCancel(parent)

	return &ContextGroup{
		ctx:    ctx,
		cancel: cancel,
	}
}

// Go calls the given function in a new goroutine.
func (cg *ContextGroup) Go(f func(context.Context) error) {
	cg.wg.Add(1)

	go func() {
		defer cg.wg.Done()

		if err := f(cg.ctx); err != nil {
			cg.errOnce.Do(func() {
				cg.err = err
				cg.cancel()
			})
		}
	}()
}

// Cancel cancels all goroutines in the group.
func (cg *ContextGroup) Cancel() {
	cg.cancel()
}

// WaitErr blocks until all function calls have returned.
// Returns the first non-nil error (if any).
func (cg *ContextGroup) WaitErr() error {
	cg.wg.Wait()
	cg.cancel()
	return cg.err
}
