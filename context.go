package synx

import (
	"context"
	"time"
)

// ContextFromSignal returns a context based on <-chan struct{}.
func ContextFromSignal(c <-chan struct{}) context.Context {
	return chanCtx(c)
}

type chanCtx <-chan struct{}

func (c chanCtx) Done() <-chan struct{}         { return c }
func (chanCtx) Deadline() (time.Time, bool)     { return time.Time{}, false }
func (c chanCtx) Value(interface{}) interface{} { return nil }
func (c chanCtx) Err() error {
	select {
	case <-c.Done():
		return context.Canceled
	default:
		return nil
	}
}

// WithCancel is a shortcut for context.WithCancel with context.Background as parent.
func WithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.Background())
}

// WithDeadline is a shortcut for context.WithDeadline with context.Background as parent.
func WithDeadline(d time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(context.Background(), d)
}

// WithTimeout is a shortcut for context.WithTimeout with context.Background as parent.
func WithTimeout(timeout time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), timeout)
}
