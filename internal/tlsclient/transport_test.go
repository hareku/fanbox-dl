package tlsclient

import (
	"net/http"
	"net/url"
	"testing"

	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/stretchr/testify/require"
)

func TestTransportUsesProxyFromEnvironment(t *testing.T) {
	t.Setenv("HTTP_PROXY", "http://http-proxy.invalid:8080")
	t.Setenv("HTTPS_PROXY", "http://https-proxy.invalid:8443")
	t.Setenv("NO_PROXY", "direct.example")
	t.Setenv("REQUEST_METHOD", "")

	transport, err := NewTransportWithOptions(tls_client.NewNoopLogger())
	require.NoError(t, err)
	t.Cleanup(transport.CloseIdleConnections)

	tests := []struct {
		url       string
		wantProxy string
	}{
		{url: "http://api.example/path", wantProxy: "http://http-proxy.invalid:8080"},
		{url: "https://api.example/path", wantProxy: "http://https-proxy.invalid:8443"},
		{url: "https://direct.example/path", wantProxy: ""},
	}

	for _, tt := range tests {
		req, reqErr := http.NewRequest(http.MethodGet, tt.url, nil)
		require.NoError(t, reqErr)
		proxyURL, proxyErr := transport.proxyFor(req)
		require.NoError(t, proxyErr)
		if tt.wantProxy == "" {
			require.Nil(t, proxyURL)
		} else {
			require.Equal(t, tt.wantProxy, proxyURL.String())
		}
	}
}

func TestTransportCachesClientByProxy(t *testing.T) {
	transport, err := NewTransportWithOptions(tls_client.NewNoopLogger())
	require.NoError(t, err)
	t.Cleanup(transport.CloseIdleConnections)

	proxyURL, err := url.Parse("http://proxy.invalid:8080")
	require.NoError(t, err)
	transport.proxyFor = func(*http.Request) (*url.URL, error) {
		return proxyURL, nil
	}

	req, err := http.NewRequest(http.MethodGet, "https://api.example/path", nil)
	require.NoError(t, err)
	first, err := transport.clientFor(req)
	require.NoError(t, err)
	require.Equal(t, proxyURL.String(), first.GetProxy())
	second, err := transport.clientFor(req)
	require.NoError(t, err)
	require.Equal(t, first, second)
	require.Len(t, transport.clients, 2)
}
