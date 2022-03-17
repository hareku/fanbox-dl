package fanbox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/cenkalti/backoff/v4"
)

type OfficialAPIClient struct {
	HTTPClient *http.Client
	Strategy   backoff.BackOff
}

func (c *OfficialAPIClient) Request(ctx context.Context, method string, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	req.Header.Set("Origin", "https://www.fanbox.cc") // If Origin header is not set, FANBOX returns HTTP 400 error.

	if err != nil {
		return nil, fmt.Errorf("http request building error: %w", err)
	}

	var resp *http.Response
	op := backoff.Operation(func() error {
		_resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return fmt.Errorf("http response error: %w", err)
		}
		resp = _resp
		return nil
	})

	err = backoff.Retry(op, backoff.WithContext(c.Strategy, ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
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
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("status is %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(v)
	if err != nil {
		return fmt.Errorf("json decoding error: %w", err)
	}
	return nil
}
