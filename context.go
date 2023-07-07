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

func (c chanCtx) Done() <-chan struct{}     { return c }
func (chanCtx) Deadline() (time.Time, bool) { return time.Time{}, false }
func (c chanCtx) Value(any) any             { return nil }
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

// used as a context key in WithValue and GetValue.
type ctxKey[T any] struct{}

// WithValue attaches val to the context. Use GetValue to get this value back.
func WithValue[T any](ctx context.Context, val T) context.Context {
	return context.WithValue(ctx, ctxKey[T]{}, val)
}

// GetValue returns the value T attached to the context.
// If there is no value attached, the zero value is returned.
func GetValue[T any](ctx context.Context) T {
	val, _ := ctx.Value(ctxKey[T]{}).(T)
	return val
}

// ContextWithoutValues returns context without any value set. However new values can be added.
func ContextWithoutValues(ctx context.Context) context.Context {
	return &contextWithoutValues{ctx}
}

type contextWithoutValues struct{ context.Context }

func (c *contextWithoutValues) Value(_ any) any { return nil }

// ContextWithValues creates a new context based on ctx and map of values.
// Shorter version of context.WithValue in a loop.
func ContextWithValues(ctx context.Context, values map[any]any) context.Context {
	return &multiValuesCtx{
		Context: ctx,
		values:  values,
	}
}

type multiValuesCtx struct {
	context.Context
	values map[any]any
}

func (c *multiValuesCtx) Value(key any) any {
	if val, ok := c.values[key]; ok {
		return val
	}
	return c.Context.Value(key)
}
