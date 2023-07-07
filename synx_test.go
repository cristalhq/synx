package synx

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestWaitSuccess(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		time.Sleep(testDelay)
		wg.Done()
	}()

	ctx := context.Background()
	if err := Wait(ctx, wg.Wait); err != nil {
		t.Fatal(err)
	}
}

func TestWaitFail(t *testing.T) {
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		time.Sleep(3 * testDelay)
		wg.Done()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(testDelay)
		cancel()
	}()

	if err := Wait(ctx, wg.Wait); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}
