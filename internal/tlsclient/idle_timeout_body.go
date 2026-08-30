package tlsclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync/atomic"
	"time"
)

var ErrRequestIdleTimeout = errors.New("request idle timeout")

type idleTimeoutBody struct {
	io.ReadCloser
	timeout time.Duration
	cancel  context.CancelFunc
}

func newIdleTimeoutBody(body io.ReadCloser, timeout time.Duration, cancel context.CancelFunc) io.ReadCloser {
	return &idleTimeoutBody{
		ReadCloser: body,
		timeout:    timeout,
		cancel:     cancel,
	}
}

func (b *idleTimeoutBody) Read(p []byte) (int, error) {
	var timedOut atomic.Bool
	timerDone := make(chan struct{})
	timer := time.AfterFunc(b.timeout, func() {
		timedOut.Store(true)
		b.cancel()
		_ = b.ReadCloser.Close()
		close(timerDone)
	})

	n, err := b.ReadCloser.Read(p)
	if !timer.Stop() {
		<-timerDone
	}
	if timedOut.Load() {
		return n, newRequestIdleTimeoutError(b.timeout)
	}

	return n, err
}

func (b *idleTimeoutBody) Close() error {
	b.cancel()
	return b.ReadCloser.Close()
}

func newRequestIdleTimeoutError(timeout time.Duration) error {
	return fmt.Errorf("%w after %s", ErrRequestIdleTimeout, timeout)
}
