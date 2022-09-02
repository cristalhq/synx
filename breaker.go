package synx

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"
)

var ErrBreakerOpen = errors.New("circuit breaker is open")

// Breaker is an implementation of Circuit Breaker pattern.
type Breaker struct {
	cfg       *BreakerConfig
	state     atomic.Value
	successes int32
	fails     int32
}

// BreakerConfig represents Breaker config.
type BreakerConfig struct {
	// Resolution is time how often we update breaker state.
	// Default is 0 which is treated as 1 second.
	Resolution time.Duration

	// FailRatio defines when breaker switches Closed -> Open (or HalfOpen if Flexible).
	// Value must be in range [0, 1] (including both border values).
	FailRatio float64

	// HalfOpenFailRatio defines when breaker switches HalfOpen -> Closed.
	// Value must be in range [0, 1] (including both border values).
	HalfOpenFailRatio float64

	// HalfOpenAllowRatio defines how many operations should be allowed in HalfOpen state.
	// Value must be in range [0, 1] (including both border values).
	HalfOpenAllowRatio float64

	// Flexible set to true allows switches Close state to HalfOpen instead of Open on a high error rate.
	// In the original circuit breaker design Close switches to Open only.
	// Default is false.
	Flexible bool
}

// Validate the config.
func (cfg *BreakerConfig) Validate() error {
	if cfg == nil {
		return errors.New("BreakerConfig cannot be nil")
	}
	if cfg.Resolution == 0 {
		cfg.Resolution = time.Second
	}
	if cfg.FailRatio < 0 || cfg.FailRatio > 1 {
		return fmt.Errorf("FailPercent must be between 0 and 1, got: %v", cfg.FailRatio)
	}
	if cfg.HalfOpenFailRatio < 0 || cfg.HalfOpenFailRatio > 1 {
		return fmt.Errorf("HalfOpenFailRatio must be between 0 and 1, got: %v", cfg.HalfOpenFailRatio)
	}
	if cfg.HalfOpenAllowRatio < 0 || cfg.HalfOpenAllowRatio > 1 {
		return fmt.Errorf("HalfOpenAllowRatio must be between 0 and 1, got: %v", cfg.HalfOpenAllowRatio)
	}
	return nil
}

// BreakerState representa one of 3 possible states of breaker.
type BreakerState int

const (
	BreakerStateOpen BreakerState = iota
	BreakerStateHalfOpen
	BreakerStateClosed
)

// String representation of the Breaker state.
func (s BreakerState) String() string {
	switch s {
	case BreakerStateOpen:
		return "open"
	case BreakerStateHalfOpen:
		return "half-open"
	case BreakerStateClosed:
		return "closed"
	default:
		return fmt.Sprintf("unknown state: %d", s)
	}
}

// NewBreaker returns new Breaker.
func NewBreaker(cfg *BreakerConfig) (*Breaker, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	cb := &Breaker{
		cfg: cfg,
	}
	cb.toState(BreakerStateClosed, time.Now().UnixNano())

	return cb, nil
}

// State of the breaker.
func (cb *Breaker) State() BreakerState {
	return cb.state.Load().(*state).curr
}

// Do the given action if breaker allows.
func (cb *Breaker) Do(fn func() error) error {
	if !cb.Allow() {
		return ErrBreakerOpen
	}

	var err error
	defer cb.Done(err == nil)
	err = fn()
	return err
}

// Done informs breaker about operation success.
// Must be used after Allow method. See examples.
func (cb *Breaker) Done(success bool) {
	if success {
		atomic.AddInt32(&cb.successes, 1)
	} else {
		atomic.AddInt32(&cb.fails, 1)
	}
}

// Allow return true when action is allwed by breaker.
// If returns true, Done method must be used after the operation. See examples.
func (cb *Breaker) Allow() bool {
	now := time.Now().UnixNano()
	state := *cb.state.Load().(*state)

	if now <= state.untilTime {
		return (state.curr != BreakerStateOpen) &&
			(state.curr == BreakerStateClosed ||
				rand.Float64() < cb.cfg.HalfOpenAllowRatio)
	}
	return cb.doAllow(state, now)
}

func (cb *Breaker) doAllow(state state, now int64) bool {
	if state.curr == BreakerStateOpen {
		cb.toState(BreakerStateHalfOpen, now)
		return true
	}

	successes, fails := atomic.LoadInt32(&cb.successes), atomic.LoadInt32(&cb.fails)
	total := int64(successes + fails)
	if total == 0 {
		cb.toState(BreakerStateClosed, now)
		return true
	}

	failRate := float64(fails) / float64(total)
	newState := BreakerStateOpen

	var ok bool
	if state.curr == BreakerStateHalfOpen {
		ok = failRate < cb.cfg.HalfOpenFailRatio
	} else {
		ok = failRate < cb.cfg.FailRatio
		if !ok && cb.cfg.Flexible {
			newState = BreakerStateHalfOpen
		}
	}

	if ok {
		newState = BreakerStateClosed
	}
	cb.toState(newState, now)
	return ok
}

func (cb *Breaker) toState(newState BreakerState, now int64) {
	atomic.StoreInt32(&cb.successes, 0)
	atomic.StoreInt32(&cb.fails, 0)

	cb.state.Store(&state{
		curr:      newState,
		untilTime: now + int64(cb.cfg.Resolution),
	})
}

type state struct {
	curr      BreakerState
	untilTime int64
}
