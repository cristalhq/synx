package synx

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const infiniteTimeout = 30 * 24 * time.Hour // domain specific infinite

// HedgedWorker ...
type HedgedWorker interface {
	Execute(ctx context.Context, input any) (result any, err error)
}

// NewHedger returns a new http.RoundTripper which implements hedged requests pattern.
// Given RoundTripper starts a new request after a timeout from previous request.
// Starts no more than upto requests.
func NewHedger(timeout time.Duration, upto int, worker HedgedWorker) HedgedWorker {
	switch {
	case timeout < 0:
		panic("synx: timeout cannot be negative")
	case upto < 1:
		panic("synx: upto must be greater than 0")
	case worker == nil:
		panic("synx: worker cannot be nil")
	}

	if timeout == 0 {
		timeout = time.Nanosecond // smallest possible timeout if not set
	}

	hedged := &hedgedWorker{
		worker:  worker,
		timeout: timeout,
		upto:    upto,
		wp:      NewWorkerPool(10, time.Minute),
	}
	return hedged
}

type hedgedWorker struct {
	worker  HedgedWorker
	timeout time.Duration
	upto    int
	wp      *WorkerPool
}

func (ht *hedgedWorker) Execute(ctx context.Context, input any) (any, error) {
	mainCtx := ctx

	var timeout time.Duration
	errOverall := &MultiError{}
	resultCh := make(chan indexedResult, ht.upto)
	errorCh := make(chan error, ht.upto)

	resultIdx := -1
	cancels := make([]func(), ht.upto)

	defer ht.wp.Do(func() {
		for i, cancel := range cancels {
			if i != resultIdx && cancel != nil {
				cancel()
			}
		}
	})

	for sent := 0; len(errOverall.Errors) < ht.upto; sent++ {
		if sent < ht.upto {
			idx := sent
			subCtx, cancel := context.WithCancel(ctx)
			cancels[idx] = cancel

			ht.wp.Do(func() {
				result, err := ht.worker.Execute(subCtx, input)
				if err != nil {
					errorCh <- err
				} else {
					resultCh <- indexedResult{idx, result}
				}
			})
		}

		// all request sent - effectively disabling timeout between requests
		if sent == ht.upto {
			timeout = infiniteTimeout
		}
		result, err := waitResult(mainCtx, resultCh, errorCh, timeout)

		switch {
		case result.Result != nil:
			resultIdx = result.Index
			return result.Result, nil
		case mainCtx.Err() != nil:
			return nil, mainCtx.Err()
		case err != nil:
			errOverall.Errors = append(errOverall.Errors, err)
		}
	}

	// all request have returned errors
	return nil, errOverall
}

func waitResult(ctx context.Context, resultCh <-chan indexedResult, errorCh <-chan error, timeout time.Duration) (indexedResult, error) {
	// try to read result first before blocking on all other channels
	select {
	case res := <-resultCh:
		return res, nil
	default:
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		select {
		case res := <-resultCh:
			return res, nil

		case err := <-errorCh:
			return indexedResult{}, err

		case <-ctx.Done():
			return indexedResult{}, ctx.Err()

		case <-timer.C:
			return indexedResult{}, nil // it's not a request timeout, it's timeout BETWEEN consecutive requests
		}
	}
}

type indexedResult struct {
	Index  int
	Result any
}

// MultiError is an error type to track multiple errors. This is used to
// accumulate errors in cases and return them as a single "error".
// Insiper by https://github.com/hashicorp/go-multierror
type MultiError struct {
	Errors        []error
	ErrorFormatFn ErrorFormatFunc
}

func (e *MultiError) Error() string {
	fn := e.ErrorFormatFn
	if fn == nil {
		fn = listFormatFunc
	}
	return fn(e.Errors)
}

func (e *MultiError) String() string {
	return fmt.Sprintf("*%#v", e.Errors)
}

// ErrorOrNil returns an error if there are some.
func (e *MultiError) ErrorOrNil() error {
	switch {
	case e == nil || len(e.Errors) == 0:
		return nil
	default:
		return e
	}
}

// ErrorFormatFunc is called by MultiError to return the list of errors as a string.
type ErrorFormatFunc func([]error) string

func listFormatFunc(es []error) string {
	if len(es) == 1 {
		return fmt.Sprintf("1 error occurred:\n\t* %s\n\n", es[0])
	}

	points := make([]string, len(es))
	for i, err := range es {
		points[i] = fmt.Sprintf("* %s", err)
	}

	return fmt.Sprintf("%d errors occurred:\n\t%s\n\n", len(es), strings.Join(points, "\n\t"))
}
