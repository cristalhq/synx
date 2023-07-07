package synx_test

import (
	"time"

	"github.com/cristalhq/synx"
)

func ExampleBlockForever() {
	ch := make(chan struct{})
	go func() {
		synx.BlockForever()
		close(ch)
	}()

	select {
	case <-time.After(50 * time.Millisecond):
	case <-ch:
		panic("goroutine must be blocked")
	}

	// Output:
}

func ExampleAsync() {
	done := synx.Async(func() {
		time.Sleep(10 * time.Millisecond)
	})

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		panic("timeout")
	}
}

func ExampleDrain() {
	ch := make(chan int)

	go func() {
		for i := 0; i < 10; i++ {
			ch <- i
		}
		close(ch)
	}()

	synx.Drain(ch)

	// Output:
}
