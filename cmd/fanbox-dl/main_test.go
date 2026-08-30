package main

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/hareku/fanbox-dl/internal/tlsclient"
	"github.com/stretchr/testify/require"
)

func TestCheckRetryDoesNotRetryRequestIdleTimeout(t *testing.T) {
	retry, err := checkRetry(
		context.Background(),
		nil,
		fmt.Errorf("request failed: %w", tlsclient.ErrRequestIdleTimeout),
	)

	require.NoError(t, err)
	require.False(t, retry)
}

func TestRequestIdleTimeout(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		got, err := parseRequestIdleTimeout(30)
		require.NoError(t, err)
		require.Equal(t, 30*time.Second, got)
	})

	t.Run("maximum", func(t *testing.T) {
		got, err := parseRequestIdleTimeout(maxRequestIdleTimeoutSeconds)
		require.NoError(t, err)
		require.Equal(t, time.Duration(maxRequestIdleTimeoutSeconds)*time.Second, got)
	})

	t.Run("overflow", func(t *testing.T) {
		_, err := parseRequestIdleTimeout(maxRequestIdleTimeoutSeconds + 1)
		require.EqualError(t, err, "--request-idle-timeout must not exceed 9223372036 seconds")
	})
}

func TestCheckRetryRetainsDefaultHandlingForOtherErrors(t *testing.T) {
	retry, err := checkRetry(context.Background(), nil, errors.New("connection failed"))

	require.NoError(t, err)
	require.True(t, retry)
}

func TestCheckRetryPreservesRequestCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	retry, err := checkRetry(ctx, nil, tlsclient.ErrRequestIdleTimeout)

	require.ErrorIs(t, err, context.Canceled)
	require.False(t, retry)
}
