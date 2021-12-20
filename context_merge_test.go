package synx

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestContextMergeCancel1(t *testing.T) {
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx := ContextMerge(ctx1, context.Background())
	cancel1()

	waitFor(t, ctx.Done())

	if err := ctx.Err(); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want %v", err, context.Canceled)
	}
}

func TestContextMergeCancel2(t *testing.T) {
	ctx2, cancel2 := context.WithCancel(context.Background())
	ctx := ContextMerge(context.Background(), ctx2)
	cancel2()

	waitFor(t, ctx.Done())

	if err := ctx.Err(); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want %v", err, context.Canceled)
	}
}

func TestContextMergeCancelBoth1(t *testing.T) {
	ctx1, cancel1 := context.WithCancel(context.Background())
	ctx2, cancel2 := context.WithCancel(context.Background())
	defer cancel2()

	ctx := ContextMerge(ctx1, ctx2)
	cancel1()

	waitFor(t, ctx.Done())

	if err := ctx.Err(); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want %v", err, context.Canceled)
	}
}

func TestContextMergeCancelBoth2(t *testing.T) {
	ctx1, cancel1 := context.WithCancel(context.Background())
	defer cancel1()

	ctx2, cancel2 := context.WithCancel(context.Background())

	ctx := ContextMerge(ctx1, ctx2)
	cancel2()

	waitFor(t, ctx.Done())

	if err := ctx.Err(); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want %v", err, context.Canceled)
	}
}

func TestErrNoErr(t *testing.T) {
	ctx := ContextMerge(context.Background(), context.Background())
	if err := ctx.Err(); !errors.Is(err, nil) {
		t.Fatalf("got %v, want %v", err, nil)
	}
}

func TestDeadline1(t *testing.T) {
	tt := time.Now().Add(5 * time.Second).UTC()
	ctx1, cancel1 := context.WithDeadline(context.Background(), tt)
	defer cancel1()

	ctx := ContextMerge(ctx1, context.Background())

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal()
	}
	if !tt.Equal(deadline) {
		t.Fatal()
	}
}

func TestDeadline2(t *testing.T) {
	tt := time.Now().Add(5 * time.Second).UTC()
	ctx1, cancel1 := context.WithDeadline(context.Background(), tt)
	defer cancel1()

	ctx := ContextMerge(context.Background(), ctx1)

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal()
	}
	if !tt.Equal(deadline) {
		t.Fatal()
	}
}

func TestDeadlineBoth1(t *testing.T) {
	t1 := time.Now().Add(5 * time.Second).UTC()
	ctx1, cancel1 := context.WithDeadline(context.Background(), t1)
	defer cancel1()

	t2 := time.Now().Add(10 * time.Second).UTC()
	ctx2, cancel2 := context.WithDeadline(context.Background(), t2)
	defer cancel2()

	ctx := ContextMerge(ctx1, ctx2)

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal()
	}
	if !t1.Equal(deadline) {
		t.Fatal()
	}
}

func TestDeadlineBoth2(t *testing.T) {
	t1 := time.Now().Add(5 * time.Second).UTC()
	ctx1, cancel1 := context.WithDeadline(context.Background(), t1)
	defer cancel1()

	t2 := time.Now().Add(10 * time.Second).UTC()
	ctx2, cancel2 := context.WithDeadline(context.Background(), t2)
	defer cancel2()

	ctx := ContextMerge(ctx2, ctx1)

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal()
	}
	if !t1.Equal(deadline) {
		t.Fatal()
	}
}

func TestValue(t *testing.T) {
	ctx1 := context.WithValue(context.Background(), "foo", "bar")
	ctx2 := context.WithValue(context.Background(), "baz", "qux")

	ctx := ContextMerge(ctx1, ctx2)

	if val := ctx.Value("foo"); val != "bar" {
		t.Fatalf("got %v, want %v", val, "bar")
	}
	if val := ctx.Value("baz"); val != "qux" {
		t.Fatalf("got %v, want %v", val, "qux")
	}
}

func TestDoneRace(t *testing.T) {
	ctx1, cancel1 := context.WithDeadline(context.Background(), time.Now())
	defer cancel1()

	ctx2, cancel2 := context.WithDeadline(context.Background(), time.Now())
	defer cancel2()

	ctx := ContextMerge(ctx1, ctx2)
	done := make(chan struct{})

	go func() {
		<-ctx.Done()
		done <- struct{}{}
	}()
	go func() {
		<-ctx.Done()
		done <- struct{}{}
	}()

	waitFor(t, done)
	waitFor(t, done)
}

func TestErrRace(t *testing.T) {
	// This test is designed to be run with the race detector enabled.
	ctx1, cancel1 := context.WithDeadline(context.Background(), time.Now())
	defer cancel1()

	ctx2, cancel2 := context.WithDeadline(context.Background(), time.Now())
	defer cancel2()

	ctx := ContextMerge(ctx1, ctx2)
	done := make(chan struct{})

	go func() {
		ctx.Err()
		done <- struct{}{}
	}()
	go func() {
		ctx.Err()
		done <- struct{}{}
	}()

	waitFor(t, done)
	waitFor(t, done)
}

func waitFor(t *testing.T, ch <-chan struct{}) {
	select {
	case <-ch:
		return
	case <-time.After(time.Second):
		t.Fatalf("timed out")
	}
}
