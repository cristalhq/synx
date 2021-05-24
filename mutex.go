package synx

import (
	"sync"
)

// TryMutex is a stub till this is accepted https://github.com/golang/go/issues/45435
type TryMutex struct {
	c        sync.Cond
	isLocked bool
}

func NewTryMutex() *TryMutex {
	tl := &TryMutex{}
	tl.c.L = &sync.Mutex{}
	return tl
}

func (tl *TryMutex) Lock() {
	tl.c.L.Lock()
	defer tl.c.L.Unlock()

	for tl.isLocked {
		tl.c.Wait()
	}
	tl.isLocked = true
}

func (tl *TryMutex) Unlock() {
	tl.c.L.Lock()
	defer tl.c.L.Unlock()

	tl.isLocked = false
	tl.c.Signal()
}

func (tl *TryMutex) TryLock() bool {
	tl.c.L.Lock()
	defer tl.c.L.Unlock()

	if tl.isLocked {
		return false
	}
	tl.isLocked = true
	return true
}
