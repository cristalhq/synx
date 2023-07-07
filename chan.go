package synx

// Drain the channel.
func Drain[T any](ch chan T) {
	for range ch {
	}
}

// ClosedChan of type T.
func ClosedChan[T any]() chan T {
	ch := make(chan T)
	close(ch)
	return ch
}
