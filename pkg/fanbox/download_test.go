package fanbox

import (
	"fmt"
	"testing"

	"github.com/hareku/fanbox-dl/internal/tlsclient"
	"github.com/stretchr/testify/require"
)

func TestRequestIdleTimeoutIsRetryable(t *testing.T) {
	err := fmt.Errorf("save a file: %w", tlsclient.ErrRequestIdleTimeout)
	require.True(t, isRetryableDownloadError(err))
}
