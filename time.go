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
