package tlsclient

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// ErrRequestIdleTimeout indicates that a request stopped receiving response data.
var ErrRequestIdleTimeout = errors.New("request idle timeout")

type idleTimeoutBody struct {
	io.ReadCloser
	timeout time.Duration
	cancel  context.CancelFunc

	mu        sync.Mutex
	timer     *time.Timer
	timedOut  bool
	closed    bool
	closeErr  error
	closeOnce sync.Once
}

func newIdleTimeoutBody(body io.ReadCloser, timeout time.Duration, cancel context.CancelFunc) io.ReadCloser {
	b := &idleTimeoutBody{
		ReadCloser: body,
		timeout:    timeout,
		cancel:     cancel,
	}
	b.timer = time.AfterFunc(timeout, b.triggerTimeout)
	return b
}

func (b *idleTimeoutBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)

	b.mu.Lock()
	if b.timedOut {
		b.mu.Unlock()
		return n, newRequestIdleTimeoutError(b.timeout)
	}
	if b.closed {
		b.mu.Unlock()
		return n, err
	}
	if err != nil {
		b.timer.Stop()
		b.mu.Unlock()
		return n, err
	}
	if n == 0 {
		b.mu.Unlock()
		return n, nil
	}
	if !b.timer.Stop() {
		b.mu.Unlock()
		b.triggerTimeout()
		return n, newRequestIdleTimeoutError(b.timeout)
	}
	b.timer.Reset(b.timeout)
	b.mu.Unlock()

	return n, nil
}

func (b *idleTimeoutBody) Close() error {
	b.mu.Lock()
	b.closed = true
	b.timer.Stop()
	b.mu.Unlock()

	return b.closeUnderlying()
}

func (b *idleTimeoutBody) triggerTimeout() {
	b.mu.Lock()
	if b.closed || b.timedOut {
		b.mu.Unlock()
		return
	}
	b.timedOut = true
	b.mu.Unlock()

	_ = b.closeUnderlying()
}

func (b *idleTimeoutBody) closeUnderlying() error {
	b.closeOnce.Do(func() {
		b.cancel()
		b.closeErr = b.ReadCloser.Close()
	})
	return b.closeErr
}

func newRequestIdleTimeoutError(timeout time.Duration) error {
	return fmt.Errorf("%w after %s", ErrRequestIdleTimeout, timeout)
}
