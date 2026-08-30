package tlsclient

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/stretchr/testify/require"
)

func TestTransportHasNoDefaultRequestTimeout(t *testing.T) {
	oldDefaultTimeout := tls_client.DefaultTimeoutSeconds
	tls_client.DefaultTimeoutSeconds = 1
	t.Cleanup(func() {
		tls_client.DefaultTimeoutSeconds = oldDefaultTimeout
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		time.Sleep(1250 * time.Millisecond)
		_, _ = io.WriteString(w, "complete")
	}))
	defer server.Close()

	transport, err := NewTransportWithOptions(tls_client.NewNoopLogger())
	require.NoError(t, err)

	client := &http.Client{Transport: transport}
	resp, err := client.Get(server.URL)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, resp.Body.Close()) })

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "complete", string(body))
}

func TestTransportResponseBodyIdleTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		time.Sleep(250 * time.Millisecond)
		_, _ = io.WriteString(w, "too late")
	}))
	defer server.Close()

	transport, err := NewTransportWithIdleTimeout(tls_client.NewNoopLogger(), 50*time.Millisecond)
	require.NoError(t, err)

	resp, err := (&http.Client{Transport: transport}).Get(server.URL)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, resp.Body.Close()) })

	_, err = io.ReadAll(resp.Body)
	require.ErrorIs(t, err, ErrRequestIdleTimeout)
}

func TestTransportRequestIdleTimeoutWhileWaitingForHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(250 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	transport, err := NewTransportWithIdleTimeout(tls_client.NewNoopLogger(), 50*time.Millisecond)
	require.NoError(t, err)

	_, err = (&http.Client{Transport: transport}).Get(server.URL)
	require.ErrorIs(t, err, ErrRequestIdleTimeout)
}

func TestTransportResponseBodyIdleTimeoutResetsOnProgress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		for _, chunk := range []string{"still ", "making ", "steady ", "progress"} {
			_, _ = io.WriteString(w, chunk)
			flusher.Flush()
			time.Sleep(100 * time.Millisecond)
		}
	}))
	defer server.Close()

	transport, err := NewTransportWithIdleTimeout(tls_client.NewNoopLogger(), 250*time.Millisecond)
	require.NoError(t, err)

	resp, err := (&http.Client{Transport: transport}).Get(server.URL)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, resp.Body.Close()) })

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "still making steady progress", string(body))
}

func TestTransportResponseBodyIdleTimeoutCanBeDisabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		time.Sleep(50 * time.Millisecond)
		_, _ = io.WriteString(w, "complete")
	}))
	defer server.Close()

	transport, err := NewTransportWithIdleTimeout(tls_client.NewNoopLogger(), 0)
	require.NoError(t, err)

	resp, err := (&http.Client{Transport: transport}).Get(server.URL)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, resp.Body.Close()) })

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Equal(t, "complete", string(body))
}

func TestTransportHonorsRequestContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		<-r.Context().Done()
	}))
	defer server.Close()

	transport, err := NewTransportWithOptions(tls_client.NewNoopLogger())
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	resp, err := (&http.Client{Transport: transport}).Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, resp.Body.Close()) })

	cancel()
	_, err = io.ReadAll(resp.Body)
	require.ErrorIs(t, err, context.Canceled)
}

func TestTransportResponseUsesTransactionContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "complete")
	}))
	defer server.Close()

	transport, err := NewTransportWithOptions(tls_client.NewNoopLogger())
	require.NoError(t, err)

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, server.URL, nil)
	require.NoError(t, err)
	resp, err := (&http.Client{Transport: transport}).Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, resp.Body.Close()) })

	require.NoError(t, resp.Request.Context().Err())
	require.NoError(t, resp.Body.Close())
	require.ErrorIs(t, resp.Request.Context().Err(), context.Canceled)
	require.NoError(t, req.Context().Err())
}
