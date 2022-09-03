package synx_test

import (
	"errors"
	"time"

	"github.com/cristalhq/synx"
)

func ExampleBreaker() {
	cb, err := synx.NewBreaker(&synx.BreakerConfig{
		Resolution:         10 * time.Millisecond,
		Requests:           1000,
		FailRatio:          0.25,
		HalfOpenFailRatio:  0.5,
		HalfOpenAllowRatio: 0.5,
		Flexible:           true,
	})
	if err != nil {
		panic(err)
	}

	if !cb.Allow() {
		panic("action cannot be done")
	}

	errAction := circuitBreakerAction()
	cb.Done(errAction == nil)

	// or just:
	errDo := cb.Do(func() error {
		// do anything you want
		return nil
	})

	if err != nil {
		if errors.Is(errDo, synx.ErrBreakerOpen) {
			panic("action cannot be done")
		} else {
			// error returned by circuitBreakerAction
			panic(err)
		}
	}

	// Output:
}

func circuitBreakerAction() error {
	// do something
	return nil
}
