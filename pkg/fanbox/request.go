package fanbox

import (
	"context"
	"fmt"
	"net/http"
)

// Request sends a request to FANBOX with credentials.
func Request(ctx context.Context, sessid string, url string) (*http.Response, error) {
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("http request building error: %w", err)
	}

	req.Header.Set("Cookie", fmt.Sprintf("FANBOXSESSID=%s", sessid))
	req.Header.Set("Origin", "https://www.fanbox.cc") // If Origin header is not set, FANBOX returns HTTP 400 error.

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http response error: %w", err)
	}

	return resp, nil
}
