package tlsclient

import (
	"context"
	"net/http"
	"sync/atomic"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

// Transport implements net/http.RoundTripper using tls-client.HttpClient
type Transport struct {
	client             tls_client.HttpClient
	requestIdleTimeout time.Duration
}

// Ensure Transport implements http.RoundTripper
var _ http.RoundTripper = (*Transport)(nil)

const DefaultRequestIdleTimeout = 30 * time.Second

// NewTransportWithOptions creates a new Transport with the given options
func NewTransportWithOptions(logger tls_client.Logger, options ...tls_client.HttpClientOption) (*Transport, error) {
	return NewTransportWithIdleTimeout(logger, DefaultRequestIdleTimeout, options...)
}

// NewTransportWithIdleTimeout creates a new Transport with the given request
// idle timeout and tls-client options. A zero timeout disables idle checks.
func NewTransportWithIdleTimeout(logger tls_client.Logger, idleTimeout time.Duration, options ...tls_client.HttpClientOption) (*Transport, error) {
	// A net/http.Transport does not impose a deadline on the entire request.
	// Match that behavior so large response bodies are not cut off by
	// tls-client's default 30-second timeout. Callers can still provide an
	// explicit timeout option, and request contexts continue to cancel requests.
	clientOptions := []tls_client.HttpClientOption{tls_client.WithTimeoutSeconds(0)}
	clientOptions = append(clientOptions, options...)

	// Ensure no redirect following for RoundTripper compatibility.
	clientOptions = append(clientOptions, tls_client.WithNotFollowRedirects())

	client, err := tls_client.NewHttpClient(logger, clientOptions...)
	if err != nil {
		return nil, err
	}

	return &Transport{
		client:             client,
		requestIdleTimeout: idleTimeout,
	}, nil
}

// RoundTrip executes a single HTTP transaction
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	request := req
	cancel := context.CancelFunc(func() {})
	var timedOut atomic.Bool
	var timer *time.Timer
	var timerDone chan struct{}
	stopTimer := func() {
		if timer != nil && !timer.Stop() {
			<-timerDone
		}
	}
	if t.requestIdleTimeout > 0 {
		ctx, cancelRequest := context.WithCancel(req.Context())
		cancel = cancelRequest
		request = req.WithContext(ctx)

		timerDone = make(chan struct{})
		timer = time.AfterFunc(t.requestIdleTimeout, func() {
			timedOut.Store(true)
			cancelRequest()
			close(timerDone)
		})
	}

	// Convert net/http.Request to fhttp.Request
	fReq, err := convertToFhttpRequest(request)
	if err != nil {
		stopTimer()
		cancel()
		return nil, err
	}

	// Execute the request
	fResp, err := t.client.Do(fReq)
	stopTimer()
	if timedOut.Load() {
		if fResp != nil && fResp.Body != nil {
			_ = fResp.Body.Close()
		}
		cancel()
		return nil, newRequestIdleTimeoutError(t.requestIdleTimeout)
	}
	if err != nil {
		cancel()
		return nil, err
	}

	// Convert fhttp.Response to net/http.Response
	resp, err := convertFromFhttpResponse(fResp, req)
	if err != nil {
		cancel()
		// Close the original response body if conversion fails
		if fResp != nil && fResp.Body != nil {
			_ = fResp.Body.Close()
		}
		return nil, err
	}
	if resp.Body != nil && t.requestIdleTimeout > 0 {
		resp.Body = newIdleTimeoutBody(resp.Body, t.requestIdleTimeout, cancel)
	} else {
		cancel()
	}

	return resp, nil
}

// CloseIdleConnections closes any idle connections
func (t *Transport) CloseIdleConnections() {
	t.client.CloseIdleConnections()
}

// convertToFhttpRequest converts net/http.Request to fhttp.Request
func convertToFhttpRequest(req *http.Request) (*fhttp.Request, error) {
	// Create new fhttp.Request with the same method and URL
	fReq, err := fhttp.NewRequest(req.Method, req.URL.String(), req.Body)
	if err != nil {
		return nil, err
	}

	// Copy headers
	fReq.Header = make(fhttp.Header)
	for key, values := range req.Header {
		for _, value := range values {
			fReq.Header.Add(key, value)
		}
	}

	// Copy other fields
	fReq.Proto = req.Proto
	fReq.ProtoMajor = req.ProtoMajor
	fReq.ProtoMinor = req.ProtoMinor
	fReq.ContentLength = req.ContentLength
	fReq.TransferEncoding = req.TransferEncoding
	fReq.Close = req.Close
	fReq.Host = req.Host
	fReq.Trailer = convertTrailerToFhttp(req.Trailer)

	// Handle GetBody function
	if req.GetBody != nil {
		fReq.GetBody = req.GetBody
	}

	// Copy context
	if req.Context() != nil {
		fReq = fReq.WithContext(req.Context())
	}

	return fReq, nil
}

// convertFromFhttpResponse converts fhttp.Response to net/http.Response
func convertFromFhttpResponse(fResp *fhttp.Response, originalReq *http.Request) (*http.Response, error) {
	if fResp == nil {
		return nil, nil
	}

	// Create new net/http.Response
	resp := &http.Response{
		Status:           fResp.Status,
		StatusCode:       fResp.StatusCode,
		Proto:            fResp.Proto,
		ProtoMajor:       fResp.ProtoMajor,
		ProtoMinor:       fResp.ProtoMinor,
		Body:             fResp.Body,
		ContentLength:    fResp.ContentLength,
		TransferEncoding: fResp.TransferEncoding,
		Close:            fResp.Close,
		Uncompressed:     fResp.Uncompressed,
		Trailer:          convertTrailerFromFhttp(fResp.Trailer),
		Request:          originalReq,
	}

	// Convert headers
	resp.Header = make(http.Header)
	for key, values := range fResp.Header {
		for _, value := range values {
			resp.Header.Add(key, value)
		}
	}

	// Note: TLS connection state from fhttp/utls is not directly compatible with crypto/tls
	// This would require a conversion function if TLS state information is needed

	return resp, nil
}

// convertTrailerToFhttp converts net/http trailer to fhttp trailer
func convertTrailerToFhttp(trailer http.Header) fhttp.Header {
	if trailer == nil {
		return nil
	}

	fTrailer := make(fhttp.Header)
	for key, values := range trailer {
		for _, value := range values {
			fTrailer.Add(key, value)
		}
	}
	return fTrailer
}

// convertTrailerFromFhttp converts fhttp trailer to net/http trailer
func convertTrailerFromFhttp(fTrailer fhttp.Header) http.Header {
	if fTrailer == nil {
		return nil
	}

	trailer := make(http.Header)
	for key, values := range fTrailer {
		for _, value := range values {
			trailer.Add(key, value)
		}
	}
	return trailer
}
