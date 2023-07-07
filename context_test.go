package synx

import (
	"context"
	"testing"
	"time"
)

const testDelay = 100 * time.Millisecond

func TestContextFromSignal(t *testing.T) {
	ch := make(chan struct{})
	ctx := ContextFromSignal(ch)

	if val := ctx.Value("anything"); val != nil {
		t.Fatal("must store nothing")
	}

	deadline, ok := ctx.Deadline()
	if want := (time.Time{}); ok || !want.Equal(deadline) {
		t.Fatalf("want (%v, %v) got (%v, %v)", want, false, deadline, ok)
	}

	go func() {
		time.Sleep(testDelay)
		close(ch)
	}()

	if err := ctx.Err(); err != nil {
		t.Fatal(err)
	}

	select {
	case <-ctx.Done():
		t.Fatal("must not be done yet")
	default:
		// pass
	}

	<-ch

	select {
	case <-ctx.Done():
	default:
		t.Fatal("must be done already")
	}

	if err := ctx.Err(); err == nil {
		t.Fatal("must be nil")
	}
}

func TestWithCancel(t *testing.T) {
	ctx, cancel := WithCancel()

	if err := ctx.Err(); err != nil {
		t.Fatal(err)
	}

	cancel()

	select {
	case <-ctx.Done():
	default:
		t.Fatal()
	}
}

func TestWithDeadline(t *testing.T) {
	ctx, cancel := WithDeadline(time.Now().Add(100 * time.Millisecond))
	defer cancel()

	if err := ctx.Err(); err != nil {
		t.Fatal(err)
	}

	select {
	case <-ctx.Done():
	case <-time.After(150 * time.Millisecond):
		t.Fatal()
	}
}

func TestWithTimeout(t *testing.T) {
	ctx, cancel := WithTimeout(100 * time.Millisecond)
	defer cancel()

	if err := ctx.Err(); err != nil {
		t.Fatal(err)
	}

	select {
	case <-ctx.Done():
	case <-time.After(150 * time.Millisecond):
		t.Fatal()
	}
}

func TestContextWithoutValues(t *testing.T) {
	ctx := context.WithValue(context.Background(), "foo", "bar")
	ctx = ContextWithoutValues(ctx)

	if got := ctx.Value("foo"); got != nil {
		t.Fatalf("got %v, want %v", got, nil)
	}
}

func TestContextWithValues(t *testing.T) {
	ctx := context.WithValue(context.Background(), "foo", "bar")
	ctx = ContextWithValues(ctx, map[any]any{
		"a": "b",
		10:  20,
	})

	if got := ctx.Value("foo"); got != "bar" {
		t.Fatalf("got %v, want %v", got, "bar")
	}
	if got := ctx.Value("a"); got != "b" {
		t.Fatalf("got %v, want %v", got, nil)
	}
	if got := ctx.Value(10); got != 20 {
		t.Fatalf("got %v, want %v", got, nil)
	}
}
