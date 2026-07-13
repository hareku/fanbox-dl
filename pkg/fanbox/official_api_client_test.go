//go:build integration

package fanbox_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
	"github.com/hareku/fanbox-dl/internal/tlsclient"
	"github.com/hareku/fanbox-dl/pkg/fanbox"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateSession_Valid(t *testing.T) {
	sessID := os.Getenv("FANBOXSESSID")
	if sessID == "" {
		t.Skip("FANBOXSESSID env var is not set")
	}

	c := newClient(t, fmt.Sprintf("FANBOXSESSID=%s", sessID))
	require.NoError(t, c.ValidateSession(context.Background()))
}

func TestValidateSession_Missing(t *testing.T) {
	c := newClient(t, "")
	err := c.ValidateSession(context.Background())
	require.Error(t, err)
	assert.True(t, errors.Is(err, fanbox.ErrInvalidSession))
}

func TestValidateSession_Invalid(t *testing.T) {
	t.Setenv("FANBOXSESSID", "definitely_invalid_value")

	sessID := os.Getenv("FANBOXSESSID")
	require.NotEmpty(t, sessID)

	c := newClient(t, fmt.Sprintf("FANBOXSESSID=%s", sessID))
	err := c.ValidateSession(context.Background())
	require.Error(t, err)
}

func newClient(t *testing.T, cookie string) *fanbox.OfficialAPIClient {
	t.Helper()

	httpClient := retryablehttp.NewClient()
	httpClient.HTTPClient.Jar = fanbox.NewCookieJar()
	tlsTransp, err := tlsclient.NewTransportWithOptions(tls_client.NewNoopLogger(), tls_client.WithClientProfile(profiles.Chrome_146_PSK))
	require.NoError(t, err)
	httpClient.HTTPClient.Transport = tlsTransp

	return &fanbox.OfficialAPIClient{
		HTTPClient: httpClient,
		Cookie:     cookie,
		UserAgent:  "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/146.0.0.0 Safari/537.36",
	}
}
