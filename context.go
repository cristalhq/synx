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
