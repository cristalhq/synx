package synx

import (
	"context"
	"sync"
	"time"
)

// ContextMerge two conext into one, with values from both contexts, with a earliest deadline.
// ctx1 is preferred for Value and Err.
//
// See: https://github.com/golang/go/issues/36503
func ContextMerge(ctx1, ctx2 context.Context) context.Context {
	ctx := &mergedContext{
		ctx1:  ctx1,
		ctx2:  ctx2,
		done1: ctx1.Done(),
		done2: ctx2.Done(),
	}
	switch {
	case ctx.done1 != nil && ctx.done2 == nil:
		ctx.done = ctx.done1
	case ctx.done1 == nil && ctx.done2 != nil:
		ctx.done = ctx.done2
	}
	return ctx
}

type mergedContext struct {
	ctx1, ctx2   context.Context
	done1, done2 <-chan struct{}

	doneOnce sync.Once
	done     <-chan struct{}

	errOnce sync.Once
	err     error
}

// Deadline implements context.Context
func (ctx *mergedContext) Deadline() (deadline time.Time, ok bool) {
	d1, ok1 := ctx.ctx1.Deadline()
	d2, ok2 := ctx.ctx2.Deadline()
	switch {
	case ok1 && ok2:
		if d1.Before(d2) {
			return d1, true
		}
		return d2, true
	case ok1:
		return d1, true
	default:
		return d2, ok2
	}
}

// Done implements context.Context
func (ctx *mergedContext) Done() <-chan struct{} {
	if ctx.done1 == nil || ctx.done2 == nil {
		return ctx.done
	}
	ctx.doneOnce.Do(func() {
		done := make(chan struct{})
		ctx.done = done
		go func() {
			select {
			case <-ctx.done1:
			case <-ctx.done2:
			}
			close(done)
		}()
	})
	return ctx.done
}

// Err implements context.Context
func (ctx *mergedContext) Err() error {
	err1 := ctx.ctx1.Err()
	err2 := ctx.ctx2.Err()
	if err1 == nil && err2 == nil {
		return nil
	}
	ctx.errOnce.Do(func() {
		if err1 != nil {
			ctx.err = err1
		} else {
			ctx.err = err2
		}
	})
	return ctx.err
}

// Value implements context.Context
func (ctx *mergedContext) Value(key interface{}) interface{} {
	if v := ctx.ctx1.Value(key); v != nil {
		return v
	}
	return ctx.ctx2.Value(key)
}
