package parser

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"
)

type blockingReadCloser struct{ closed chan struct{} }

func (r *blockingReadCloser) Read([]byte) (int, error) { <-r.closed; return 0, io.ErrClosedPipe }
func (r *blockingReadCloser) Close() error {
	select {
	case <-r.closed:
	default:
		close(r.closed)
	}
	return nil
}

func TestParseWithReadCloserCancellation(t *testing.T) {
	r := &blockingReadCloser{closed: make(chan struct{})}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	_, err := ParseWithReadCloser(ctx, r, Options{})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err=%v", err)
	}
}
