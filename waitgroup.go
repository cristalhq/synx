package synx

import (
	"sync"
)

// WaitGroup is like sync.WaitGroup with a signal channel.
type WaitGroup struct {
	wg     sync.WaitGroup
	doneCh chan struct{}
}

// NewWaitGroup returns a new WaitGroup.
func NewWaitGroup() *WaitGroup {
	return &WaitGroup{
		doneCh: make(chan struct{}),
	}
}

// Go run the given fn guarded with a wait group.
func (wg *WaitGroup) Go(fn func()) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		fn()
	}()
}

// Add has same behaviour as sync.WaitGroup.
func (wg *WaitGroup) Add(delta int) {
	wg.wg.Add(delta)
}

// Done has same behaviour as sync.WaitGroup.
func (wg *WaitGroup) Done() {
	wg.wg.Done()
}

// Wait has same behaviour as sync.WaitGroup.
func (wg *WaitGroup) Wait() {
	wg.wg.Wait()
	close(wg.doneCh)
}

// DoneChan returns a channel that will be closed on completion.
// Note: Wait method must be executed by the user to make it work.
func (wg *WaitGroup) DoneChan() <-chan struct{} {
	return wg.doneCh
}
