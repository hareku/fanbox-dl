package fanbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
)

type ApiClient interface {
	// Request sends a request to the URL with credentials.
	Request(ctx context.Context, url string) (*http.Response, error)

	// RequestAsJSON requests with credentials, and unmarshal the response body as v.
	RequestAsJSON(ctx context.Context, url string, v interface{}) error
}

type httpApiClient struct {
	sessionID string
}

func NewHttpApiClient(sessionID string) ApiClient {
	return &httpApiClient{sessionID}
}

// Request sends a request to the specified FANBOX URL with credentials.
func (c *httpApiClient) Request(ctx context.Context, url string) (*http.Response, error) {
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("http request building error: %w", err)
	}

	req.Header.Set("Cookie", fmt.Sprintf("FANBOXSESSID=%s", c.sessionID))
	req.Header.Set("Origin", "https://www.fanbox.cc") // If Origin header is not set, FANBOX returns HTTP 400 error.

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http response error: %w", err)
	}

	return resp, nil
}

func (c *httpApiClient) RequestAsJSON(ctx context.Context, url string, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("v of RequestAsJSON should be a pointer")
	}

	resp, err := c.Request(ctx, url)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("body reading error: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("status code is %d, response body: %s", resp.StatusCode, body)
	}

	err = json.Unmarshal(body, v)
	if err != nil {
		return fmt.Errorf("json unmarshal error: %w", err)
	}

	return nil
}
