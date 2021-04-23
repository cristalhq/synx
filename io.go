package synx

import (
	"context"
	"io"
)

type readerFn func(p []byte) (n int, err error)

func (rf readerFn) Read(p []byte) (n int, err error) {
	return rf(p)
}

// Copy with cancellation.
func Copy(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	readerCtx := readerFn(func(p []byte) (int, error) {
		select {
		case <-ctx.Done():
			return 0, ctx.Err()
		default:
			return src.Read(p)
		}
	})

	written, err := io.Copy(dst, readerCtx)
	return written, err
}
