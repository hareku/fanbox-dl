package main

import (
	"context"
	"errors"
	"fmt"
	"testing"

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
