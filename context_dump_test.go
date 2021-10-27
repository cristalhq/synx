package synx

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestDumpContext(t *testing.T) {
	withCancel := func(ctx context.Context) context.Context {
		ctx, _ = context.WithCancel(ctx)
		return ctx
	}

	withTimeout := func(ctx context.Context) context.Context {
		ctx, _ = context.WithTimeout(ctx, time.Second)
		return ctx
	}

	withDeadline := func(ctx context.Context) context.Context {
		ctx, _ = context.WithDeadline(ctx, time.Now())
		return ctx
	}

	withNop := func(ctx context.Context) context.Context {
		type nopCtx struct {
			context.Context
		}
		return nopCtx{ctx}
	}

	testCases := []struct {
		ctx        context.Context
		wantValues map[interface{}]interface{}
	}{
		{
			ctx:        nil,
			wantValues: nil,
		},
		{
			ctx:        context.Background(),
			wantValues: map[interface{}]interface{}{},
		},
		{
			ctx:        withCancel(context.Background()),
			wantValues: map[interface{}]interface{}{},
		},
		{
			ctx:        withTimeout(context.Background()),
			wantValues: map[interface{}]interface{}{},
		},
		{
			ctx:        withDeadline(context.Background()),
			wantValues: map[interface{}]interface{}{},
		},
		{
			ctx:        withNop(context.Background()),
			wantValues: map[interface{}]interface{}{},
		},
		{
			ctx: context.WithValue(context.Background(), "foo", "bar"),
			wantValues: map[interface{}]interface{}{
				"foo": "bar",
			},
		},
		{
			ctx: context.WithValue(context.WithValue(
				context.WithValue(context.Background(), "foo1", "bar1"),
				"foo2", "bar2"),
				"foo3", "bar3"),
			wantValues: map[interface{}]interface{}{
				"foo1": "bar1",
				"foo2": "bar2",
				"foo3": "bar3",
			},
		},
		{
			ctx: withDeadline(context.WithValue(
				withTimeout(context.WithValue(
					withCancel(context.WithValue(
						context.Background(), "foo", "bar"),
					), "foo2", "bar2"),
				), "foo3", "bar3"),
			),
			wantValues: map[interface{}]interface{}{
				"foo":  "bar",
				"foo2": "bar2",
				"foo3": "bar3",
			},
		},
		{
			ctx: withDeadline(context.WithValue(
				withTimeout(context.WithValue(
					withNop(withCancel(context.WithValue(
						context.Background(), "foo", "bar")),
					), "foo2", "bar2"),
				), "foo3", "bar3"),
			),
			wantValues: map[interface{}]interface{}{
				"foo2": "bar2",
				"foo3": "bar3",
			},
		},
	}

	for i, test := range testCases {
		values := DumpContext(test.ctx)
		if !reflect.DeepEqual(values, test.wantValues) {
			t.Fatalf("#%d: want %+v, got %+v", i+1, test.wantValues, values)
		}
	}
}
