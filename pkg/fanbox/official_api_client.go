package fanbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/hashicorp/go-retryablehttp"
)

type OfficialAPIClient struct {
	HTTPClient *retryablehttp.Client
	SessionID  string
}

func (c *OfficialAPIClient) Request(ctx context.Context, method string, url string) (*http.Response, error) {
	req, err := retryablehttp.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("http request building error: %w", err)
	}

	req = req.WithContext(ctx)
	req.Header.Set("Cookie", fmt.Sprintf("FANBOXSESSID=%s", c.SessionID))
	req.Header.Set("Origin", "https://www.fanbox.cc") // If Origin header is not set, FANBOX returns HTTP 400 error.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36")

	return c.HTTPClient.Do(req)
}

func (c *OfficialAPIClient) RequestAndUnwrapJSON(ctx context.Context, method string, url string, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("v should be a pointer")
	}

	resp, err := c.Request(ctx, method, url)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return fmt.Errorf("status is %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return fmt.Errorf("json decoding error: %w", err)
	}
	return nil
}
