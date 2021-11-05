package synx

import (
	"context"
	"time"
)

// Sleep with cancellation. Returns true whether timer has fired.
func Sleep(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func SlowStartTimer(ctx context.Context, d time.Duration, before ...time.Duration) (<-chan time.Time, context.CancelFunc) {
	if len(before) == 0 {
		t := time.NewTimer(d)
		return t.C, func() { t.Stop() }
	}

	ch := make(chan time.Time)
	stop := func() { close(ch) }

	go func() {
		defer stop()

		t := time.NewTimer(d)
		for _, b := range before {
			select {
			case <-ctx.Done():
				return
			case now := <-time.After(b):
				ch <- now
			}
		}

		if !t.Stop() {
			<-t.C
		}
		t.Reset(d)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case now := <-t.C:
				ch <- now
			}
		}
	}()
	return ch, stop
}
