package fanbox

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	fhttp "github.com/bogdanfinn/fhttp"
	"github.com/hashicorp/go-retryablehttp"
)

const (
	chromeTestUA  = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36"
	firefoxTestUA = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:152.0) Gecko/20100101 Firefox/152.0"
)

func TestResolveBrowserProfile(t *testing.T) {
	tests := []struct {
		name             string
		userAgent        string
		requestedProfile string
		wantFamily       BrowserFamily
		wantErr          bool
	}{
		{
			name:             "auto defaults chrome for default UA",
			userAgent:        chromeTestUA,
			requestedProfile: BrowserProfileAuto,
			wantFamily:       BrowserFamilyChrome,
		},
		{
			name:             "auto detects firefox",
			userAgent:        firefoxTestUA,
			requestedProfile: BrowserProfileAuto,
			wantFamily:       BrowserFamilyFirefox,
		},
		{
			name:             "explicit chrome overrides firefox UA",
			userAgent:        firefoxTestUA,
			requestedProfile: BrowserProfileChrome,
			wantFamily:       BrowserFamilyChrome,
		},
		{
			name:             "explicit firefox overrides chrome UA",
			userAgent:        chromeTestUA,
			requestedProfile: BrowserProfileFirefox,
			wantFamily:       BrowserFamilyFirefox,
		},
		{
			name:             "unknown auto falls back chrome",
			userAgent:        "fanbox-dl-test",
			requestedProfile: BrowserProfileAuto,
			wantFamily:       BrowserFamilyChrome,
		},
		{
			name:             "invalid errors",
			userAgent:        chromeTestUA,
			requestedProfile: "safari",
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResolveBrowserProfile(tt.userAgent, tt.requestedProfile)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ResolveBrowserProfile() error = %v", err)
			}
			if got.Family != tt.wantFamily {
				t.Fatalf("family = %q, want %q", got.Family, tt.wantFamily)
			}
		})
	}
}

func TestOfficialAPIClientRequestHeaders(t *testing.T) {
	tests := []struct {
		name             string
		requestedProfile string
		userAgent        string
		asset            bool
		assert           func(*testing.T, http.Header)
	}{
		{
			name:             "chrome api headers",
			requestedProfile: BrowserProfileChrome,
			userAgent:        chromeTestUA,
			assert: func(t *testing.T, h http.Header) {
				t.Helper()
				assertHeader(t, h, "Sec-Ch-Ua-Mobile", "?0")
				assertHeader(t, h, "Origin", "https://www.fanbox.cc")
				assertHeader(t, h, "Accept", "application/json, text/plain, */*")
				assertHeader(t, h, "Sec-Fetch-Site", "same-site")
				assertHeaderOrder(t, h)
			},
		},
		{
			name:             "firefox api headers",
			requestedProfile: BrowserProfileFirefox,
			userAgent:        firefoxTestUA,
			assert: func(t *testing.T, h http.Header) {
				t.Helper()
				assertHeader(t, h, "Origin", "https://www.fanbox.cc")
				assertHeader(t, h, "Accept-Language", "en-US,en;q=0.5")
				assertHeaderAbsent(t, h, "Sec-Ch-Ua")
				assertHeaderOrder(t, h)
			},
		},
		{
			name:             "chrome asset headers",
			requestedProfile: BrowserProfileChrome,
			userAgent:        chromeTestUA,
			asset:            true,
			assert: func(t *testing.T, h http.Header) {
				t.Helper()
				assertHeader(t, h, "Accept", "*/*")
				assertHeaderAbsent(t, h, "Origin")
				assertHeader(t, h, "Referer", "https://www.fanbox.cc/")
				assertHeaderOrder(t, h)
			},
		},
		{
			name:             "firefox asset headers omit chrome client hints",
			requestedProfile: BrowserProfileFirefox,
			userAgent:        firefoxTestUA,
			asset:            true,
			assert: func(t *testing.T, h http.Header) {
				t.Helper()
				assertHeader(t, h, "Accept", "*/*")
				assertHeaderAbsent(t, h, "Origin")
				assertHeaderAbsent(t, h, "Sec-Ch-Ua")
				assertHeaderOrder(t, h)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotHeader http.Header
			api := testAPIClient(tt.userAgent, tt.requestedProfile, func(req *http.Request) (*http.Response, error) {
				gotHeader = req.Header.Clone()
				return jsonResponse(http.StatusOK, `{"body":[]}`), nil
			})

			var err error
			if tt.asset {
				resp, reqErr := api.RequestAsset(context.Background(), http.MethodGet, "https://downloads.fanbox.cc/path")
				if reqErr == nil {
					_, _ = io.Copy(io.Discard, resp.Body)
					_ = resp.Body.Close()
				}
				err = reqErr
			} else {
				var out struct {
					Body []string `json:"body"`
				}
				err = api.RequestAndUnwrapJSON(context.Background(), http.MethodGet, "https://api.fanbox.cc/test", &out)
			}
			if err != nil {
				t.Fatalf("request error = %v", err)
			}
			tt.assert(t, gotHeader)
			assertHeader(t, gotHeader, "Cookie", "FANBOXSESSID=test")
			assertHeader(t, gotHeader, "User-Agent", tt.userAgent)
		})
	}
}

func TestOfficialAPIClientAPIForbiddenWrapsSentinel(t *testing.T) {
	api := testAPIClient(chromeTestUA, BrowserProfileChrome, func(req *http.Request) (*http.Response, error) {
		return jsonResponse(http.StatusForbidden, ""), nil
	})

	var out struct{}
	err := api.RequestAndUnwrapJSON(context.Background(), http.MethodGet, "https://api.fanbox.cc/test", &out)
	if !errors.Is(err, ErrStatusForbidden) {
		t.Fatalf("error = %v, want ErrStatusForbidden", err)
	}
}

func testAPIClient(userAgent, requestedProfile string, roundTrip func(*http.Request) (*http.Response, error)) *OfficialAPIClient {
	profile, err := ResolveBrowserProfile(userAgent, requestedProfile)
	if err != nil {
		panic(err)
	}
	client := retryablehttp.NewClient()
	client.RetryMax = 0
	client.HTTPClient.Transport = roundTripFunc(roundTrip)
	return &OfficialAPIClient{
		HTTPClient:     client,
		Cookie:         "FANBOXSESSID=test",
		UserAgent:      userAgent,
		BrowserProfile: profile,
	}
}

func jsonResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func assertHeader(t *testing.T, h http.Header, key string, want string) {
	t.Helper()
	if got := h.Get(key); got != want {
		t.Fatalf("%s = %q, want %q", key, got, want)
	}
}

func assertHeaderAbsent(t *testing.T, h http.Header, key string) {
	t.Helper()
	if got := h.Get(key); got != "" {
		t.Fatalf("%s = %q, want absent", key, got)
	}
}

func assertHeaderOrder(t *testing.T, h http.Header) {
	t.Helper()
	if len(h[fhttp.HeaderOrderKey]) == 0 {
		t.Fatal("header order marker is missing")
	}
	if len(h[fhttp.PHeaderOrderKey]) == 0 {
		t.Fatal("pseudo-header order marker is missing")
	}
}
