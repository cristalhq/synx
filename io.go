package synx

import (
	"context"
	"io"
	"sync"
)

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

type readerFn func(p []byte) (n int, err error)

func (rf readerFn) Read(p []byte) (n int, err error) {
	return rf(p)
}

// SafeReader synchronizes concurrent reads to the underlying io.Writer.
func SafeReader(r io.Reader) io.Reader {
	return &safeReader{r: r}
}

type safeReader struct {
	mu sync.Mutex
	r  io.Reader
}

func (sr *safeReader) Read(b []byte) (int, error) {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	return sr.r.Read(b)
}

// SafeWriter synchronizes concurrent writes to the underlying io.Writer.
func SafeWriter(w io.Writer) io.Writer {
	return &safeWriter{w: w}
}

type safeWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (sw *safeWriter) Write(b []byte) (int, error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	return sw.w.Write(b)
}
