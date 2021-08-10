package synx

import (
	"sync"
)

// Once is wrapper around sync.Once with a signal channel.
type Once struct {
	once   sync.Once
	doneCh chan struct{}
}

// NewOnce returns a new Once.
func NewOnce() *Once {
	return &Once{
		doneCh: make(chan struct{}),
	}
}

// Do has same behaviour as sync.Once.
func (o *Once) Do(f func()) {
	o.once.Do(f)
}

// DoneChan returns a channel that will be closed on completion.
func (o *Once) DoneChan() <-chan struct{} {
	return o.doneCh
}
