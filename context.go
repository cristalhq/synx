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

// ContextWithoutValues returns context without any value set. However new values can be added.
func ContextWithoutValues(ctx context.Context) context.Context {
	return &contextWithoutValues{ctx}
}

type contextWithoutValues struct{ context.Context }

func (c *contextWithoutValues) Value(_ interface{}) interface{} { return nil }

// ContextWithValues creates a new context based on ctx and map of values.
// Shorter version of context.WithValue in a loop.
func ContextWithValues(ctx context.Context, values map[interface{}]interface{}) context.Context {
	return &multiValuesCtx{
		Context: ctx,
		values:  values,
	}
}

type multiValuesCtx struct {
	context.Context
	values map[interface{}]interface{}
}

func (c *multiValuesCtx) Value(key interface{}) interface{} {
	if val, ok := c.values[key]; ok {
		return val
	}
	return c.Context.Value(key)
}
