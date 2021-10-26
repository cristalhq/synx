package synx

import (
	"testing"
	"time"
)

const testDelay = 100 * time.Millisecond

func TestContextFromSignal(t *testing.T) {
	ch := make(chan struct{})
	ctx := ContextFromSignal(ch)

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
}
