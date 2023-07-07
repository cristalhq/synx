package synx

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestDumpContext(t *testing.T) {
	t.Skip()

	var cancel context.CancelFunc

	withCancel := func(ctx context.Context) context.Context {
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()
		return ctx
	}

	withTimeout := func(ctx context.Context) context.Context {
		ctx, cancel = context.WithTimeout(ctx, time.Second)
		defer cancel()
		return ctx
	}

	withDeadline := func(ctx context.Context) context.Context {
		ctx, cancel = context.WithDeadline(ctx, time.Now())
		defer cancel()
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
		wantValues map[any]any
	}{
		{
			ctx:        nil,
			wantValues: nil,
		},
		{
			ctx:        context.Background(),
			wantValues: map[any]any{},
		},
		{
			ctx:        withCancel(context.Background()),
			wantValues: map[any]any{},
		},
		{
			ctx:        withTimeout(context.Background()),
			wantValues: map[any]any{},
		},
		{
			ctx:        withDeadline(context.Background()),
			wantValues: map[any]any{},
		},
		{
			ctx:        withNop(context.Background()),
			wantValues: map[any]any{},
		},
		{
			ctx: context.WithValue(context.Background(), "foo", "bar"),
			wantValues: map[any]any{
				"foo": "bar",
			},
		},
		{
			ctx: context.WithValue(context.WithValue(
				context.WithValue(context.Background(), "foo1", "bar1"),
				"foo2", "bar2"),
				"foo3", "bar3"),
			wantValues: map[any]any{
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
			wantValues: map[any]any{
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
			wantValues: map[any]any{
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
