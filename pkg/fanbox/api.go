package fanbox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/cenkalti/backoff/v4"
)

//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

// API provides functions to call FANBOX APIs.
type API interface {
	// Request sends a request to URL.
	Request(ctx context.Context, method string, url string) (*http.Response, error)
	ListCreator(ctx context.Context, url string) (*ListCreator, error)
	ListPlans(ctx context.Context) (*PlanListSupporting, error)
	PostInfo(ctx context.Context, postID string) (*PostInfoBody, error)
	ListFollowing(ctx context.Context) (*CreatorListFollowing, error)
}

type webAPI struct {
	client   *http.Client
	strategy backoff.BackOff
}

func NewAPI(client *http.Client, strategy backoff.BackOff) API {
	return &webAPI{client, strategy}
}

func (w *webAPI) Request(ctx context.Context, method string, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	req.Header.Set("Origin", "https://www.fanbox.cc") // If Origin header is not set, FANBOX returns HTTP 400 error.

	if err != nil {
		return nil, fmt.Errorf("http request building error: %w", err)
	}

	var resp *http.Response
	op := backoff.Operation(func() error {
		_resp, err := w.client.Do(req)
		if err != nil {
			return fmt.Errorf("http response error: %w", err)
		}
		resp = _resp
		return nil
	})

	err = backoff.Retry(op, backoff.WithContext(w.strategy, ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (w *webAPI) ListCreator(ctx context.Context, url string) (*ListCreator, error) {
	var res ListCreator
	err := w.requestAsJSON(ctx, http.MethodGet, url, &res)
	if err != nil {
		return nil, err
	}
	return &res, err
}

func (w *webAPI) ListPlans(ctx context.Context) (*PlanListSupporting, error) {
	var res PlanListSupporting
	err := w.requestAsJSON(ctx, http.MethodGet, PlanListSupportingURL(), &res)
	if err != nil {
		return nil, err
	}
	return &res, err
}

func (w *webAPI) PostInfo(ctx context.Context, postID string) (*PostInfoBody, error) {
	var res PostInfo
	err := w.requestAsJSON(ctx, http.MethodGet, PostInfoURL(postID), &res)
	if err != nil {
		return nil, err
	}

	return &res.Body, err
}

func (w *webAPI) ListFollowing(ctx context.Context) (*CreatorListFollowing, error) {
	var res CreatorListFollowing
	err := w.requestAsJSON(ctx, http.MethodGet, CreatorListFollowingURL(), &res)
	if err != nil {
		return nil, err
	}
	return &res, err
}

func (w *webAPI) requestAsJSON(ctx context.Context, method string, url string, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("v of RequestAsJSON should be a pointer")
	}

	resp, err := w.Request(ctx, method, url)
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
