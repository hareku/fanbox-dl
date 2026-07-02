package tlsclient

import (
	"net/http"
	"testing"

	fhttp "github.com/bogdanfinn/fhttp"
)

func TestSetHeaderOrderSurvivesRequestConversion(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://api.fanbox.cc/test", nil)
	if err != nil {
		t.Fatalf("NewRequest() error = %v", err)
	}

	headerOrder := []string{"user-agent", "accept", "cookie"}
	pseudoHeaderOrder := []string{":method", ":authority", ":scheme", ":path"}
	SetHeaderOrder(req.Header, headerOrder, pseudoHeaderOrder)

	fReq, err := convertToFhttpRequest(req)
	if err != nil {
		t.Fatalf("convertToFhttpRequest() error = %v", err)
	}

	assertStringSlice(t, fReq.Header[fhttp.HeaderOrderKey], headerOrder)
	assertStringSlice(t, fReq.Header[fhttp.PHeaderOrderKey], pseudoHeaderOrder)
}

func assertStringSlice(t *testing.T, got []string, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: got %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("[%d] = %q, want %q: got %#v", i, got[i], want[i], got)
		}
	}
}
