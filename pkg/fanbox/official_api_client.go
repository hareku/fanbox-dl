package fanbox

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/http/httputil"
	"reflect"

	"github.com/hashicorp/go-retryablehttp"
)

type OfficialAPIClient struct {
	HTTPClient      *retryablehttp.Client
	AssetHTTPClient *retryablehttp.Client
	Cookie          string
	UserAgent       string
	BrowserProfile  *BrowserProfile
}

func (c *OfficialAPIClient) Request(ctx context.Context, method string, url string) (*http.Response, error) {
	return c.request(ctx, method, url, false)
}

func (c *OfficialAPIClient) RequestAsset(ctx context.Context, method string, url string) (*http.Response, error) {
	return c.request(ctx, method, url, true)
}

func (c *OfficialAPIClient) request(ctx context.Context, method string, url string, asset bool) (*http.Response, error) {
	req, err := retryablehttp.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("http request building error: %w", err)
	}

	req = req.WithContext(ctx)
	profile := c.BrowserProfile
	if profile == nil {
		profile = newChromeProfile(c.UserAgent)
	}
	if asset {
		profile.ApplyAssetHeaders(req.Header, c.Cookie, c.UserAgent)
	} else {
		profile.ApplyAPIHeaders(req.Header, c.Cookie, c.UserAgent)
	}

	httpClient := c.HTTPClient
	if asset && c.AssetHTTPClient != nil {
		httpClient = c.AssetHTTPClient
	}
	return httpClient.Do(req)
}

func (c *OfficialAPIClient) RequestAndUnwrapJSON(ctx context.Context, method string, url string, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("v should be a pointer")
	}

	resp, err := c.Request(ctx, method, url)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		if resp.StatusCode == 403 {
			return ErrStatusForbidden
		}
		return fmt.Errorf("status is %s", resp.Status)
	}

	if err = json.NewDecoder(resp.Body).Decode(v); err != nil {
		if dump, dumpErr := httputil.DumpResponse(resp, false); dumpErr == nil {
			slog.Debug("Response dump", "dump", string(dump))
		}
		return fmt.Errorf("json decoding error: %w", err)
	}
	return nil
}

var ErrFailedToThumbnailing = fmt.Errorf("failed to thumbnailing")

var ErrInvalidSession = errors.New("invalid session")

func (c *OfficialAPIClient) ValidateSession(ctx context.Context) error {
	var resp PlanListSupportingResponse
	if err := c.RequestAndUnwrapJSON(ctx, http.MethodGet, "https://api.fanbox.cc/plan.listSupporting", &resp); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidSession, err)
	}
	if len(resp.Body) == 0 {
		return ErrInvalidSession
	}
	return nil
}

// fanbox returns HTTP 500 error and response body is "failed to thumbnailing"
// when the image is not available (e.g. too large).
func IsFailedToThumbnailingErr(resp *http.Response) (bool, error) {
	if resp.StatusCode != 500 {
		return false, nil
	}
	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return false, fmt.Errorf("parse content type: %w", err)
	}
	if mediaType != "text/plain" {
		return false, nil
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("read response body: %w", err)
	}
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(b))

	if string(b) == "failed to thumbnailing" {
		return true, nil
	}
	return false, nil
}
