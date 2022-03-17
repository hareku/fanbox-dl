package fanbox

import (
	"fmt"
	"net/http"
)

func NewHTTPClientWithSession(sessionID string) *http.Client {
	return &http.Client{
		Transport: NewSessionRoundTripper(sessionID),
	}
}

func NewSessionRoundTripper(sessionID string) http.RoundTripper {
	return &sessionRoundTripper{sessionID}
}

type sessionRoundTripper struct {
	sessionID string
}

func (s *sessionRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Cookie", fmt.Sprintf("FANBOXSESSID=%s", s.sessionID))

	return http.DefaultTransport.RoundTrip(req)
}
