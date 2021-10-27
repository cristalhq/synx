package synx

import (
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
