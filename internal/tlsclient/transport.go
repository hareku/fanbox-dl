package tlsclient

import (
	"net/http"
	"net/url"
	"sync"

	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"golang.org/x/net/http/httpproxy"
)

// Transport implements net/http.RoundTripper using tls-client.HttpClient
type Transport struct {
	mu       sync.Mutex
	clients  map[string]tls_client.HttpClient
	logger   tls_client.Logger
	options  []tls_client.HttpClientOption
	proxyFor func(*http.Request) (*url.URL, error)
}

// Ensure Transport implements http.RoundTripper
var _ http.RoundTripper = (*Transport)(nil)

// NewTransportWithOptions creates a new Transport with the given options
func NewTransportWithOptions(logger tls_client.Logger, options ...tls_client.HttpClientOption) (*Transport, error) {
	// Ensure no redirect following for RoundTripper compatibility
	options = append(options, tls_client.WithNotFollowRedirects())
	proxyFromEnvironment := httpproxy.FromEnvironment().ProxyFunc()

	client, err := tls_client.NewHttpClient(logger, options...)
	if err != nil {
		return nil, err
	}

	return &Transport{
		clients: map[string]tls_client.HttpClient{"": client},
		logger:  logger,
		options: append([]tls_client.HttpClientOption(nil), options...),
		proxyFor: func(req *http.Request) (*url.URL, error) {
			return proxyFromEnvironment(req.URL)
		},
	}, nil
}

func (t *Transport) clientFor(req *http.Request) (tls_client.HttpClient, error) {
	proxyURL, err := t.proxyFor(req)
	if err != nil {
		return nil, err
	}

	proxyKey := ""
	if proxyURL != nil {
		proxyKey = proxyURL.String()
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if client, ok := t.clients[proxyKey]; ok {
		return client, nil
	}

	options := append([]tls_client.HttpClientOption(nil), t.options...)
	options = append(options, tls_client.WithProxyUrl(proxyKey))
	client, err := tls_client.NewHttpClient(t.logger, options...)
	if err != nil {
		return nil, err
	}
	t.clients[proxyKey] = client
	return client, nil
}

// RoundTrip executes a single HTTP transaction
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	client, err := t.clientFor(req)
	if err != nil {
		return nil, err
	}

	// Convert net/http.Request to fhttp.Request
	fReq, err := convertToFhttpRequest(req)
	if err != nil {
		return nil, err
	}

	// Execute the request
	fResp, err := client.Do(fReq)
	if err != nil {
		return nil, err
	}

	// Convert fhttp.Response to net/http.Response
	resp, err := convertFromFhttpResponse(fResp, req)
	if err != nil {
		// Close the original response body if conversion fails
		if fResp != nil && fResp.Body != nil {
			_ = fResp.Body.Close()
		}
		return nil, err
	}

	return resp, nil
}

// CloseIdleConnections closes any idle connections
func (t *Transport) CloseIdleConnections() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, client := range t.clients {
		client.CloseIdleConnections()
	}
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
