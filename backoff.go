package synx

import (
	"context"
	"errors"
	"time"
)

// Backoff logic for retry.
type Backoff interface {
	// Next returns a time to wait and flag to stop.
	Next() (next time.Duration, stop bool)
}

// DoWithBackoff invokes function and retry with a backoff if it is needed.
// Wrap error from function as RetryableError to continue retries.
func DoWithBackoff(ctx context.Context, b Backoff, fn func(ctx context.Context) error) error {
	for {
		err := fn(ctx)
		if err == nil || !IsRetryableError(err) {
			return err
		}

		next, stop := b.Next()
		if stop {
			return err
		}
		if err := wait(ctx, next); err != nil {
			return err
		}
	}
}

func wait(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// RetryableError wraps error as retryable.
func RetryableError(err error) error {
	if err == nil {
		return nil
	}
	return &retryableError{err: err}
}

// IsRetryableError reports whether a given error is retryable.
// Returns false for nil.
func IsRetryableError(err error) bool {
	var rerr *retryableError
	return errors.As(err, &rerr)
}

type retryableError struct {
	err error
}

func (e *retryableError) Error() string {
	if e.err == nil {
		return "retryable: <nil>"
	}
	return "retryable: " + e.err.Error()
}

func (e *retryableError) Unwrap() error {
	return e.err
}
