package fanbox

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Run_CurrentResponseBodies(t *testing.T) {
	const (
		paginationURL = "https://api.fanbox.cc/post.paginateCreator?creatorId=example"
		pageURL       = "https://api.fanbox.cc/post.listCreator?creatorId=example&limit=10"
		postInfoURL   = "https://api.fanbox.cc/post.info?postId=1001"
	)

	transport := &staticResponseTransport{responses: map[string]string{
		paginationURL: `{"body":{"pageUrls":["` + pageURL + `"]}}`,
		pageURL:       `{"body":{"posts":[{"id":"1001","title":"Example post","creatorId":"example"}]}}`,
		postInfoURL:   `{"body":{"post":{"id":"1001","title":"Example post","publishedDatetime":"2026-08-27T12:34:56+09:00","creatorId":"example","body":{"images":[{"id":"image-1","extension":"jpg","originalUrl":"https://downloads.fanbox.cc/image-1.jpg"}]}}}}`,
	}}
	client := Client{
		DryRun:            true,
		OfficialAPIClient: newStaticOfficialAPIClient(transport),
		Storage: &LocalStorage{
			SaveDir: t.TempDir(),
		},
	}

	require.NoError(t, client.Run(context.Background(), "example"))
	assert.Equal(t, []string{paginationURL, pageURL, postInfoURL}, transport.requestURLs())
}

func TestCreatorIDLister_Do_CurrentResponseBodies(t *testing.T) {
	const (
		listSupportingURL = "https://api.fanbox.cc/plan.listSupporting"
		listFollowingURL  = "https://api.fanbox.cc/creator.listFollowing"
	)

	transport := &staticResponseTransport{responses: map[string]string{
		listSupportingURL: listSupportingCurrentResponse,
		listFollowingURL:  listFollowingCurrentResponse,
	}}
	lister := CreatorIDLister{OfficialAPIClient: newStaticOfficialAPIClient(transport)}

	creatorIDs, err := lister.Do(context.Background(), &CreatorIDListerDoInput{
		IncludeSupporting: true,
		IncludeFollowing:  true,
	})
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"supported-creator", "followed-creator", "shared-creator"}, creatorIDs)
	assert.Equal(t, []string{listSupportingURL, listFollowingURL}, transport.requestURLs())
}

func TestCreatorIDLister_Do_InputCreatorIDsSkipsAutomaticListing(t *testing.T) {
	transport := &staticResponseTransport{responses: map[string]string{}}
	lister := CreatorIDLister{OfficialAPIClient: newStaticOfficialAPIClient(transport)}

	creatorIDs, err := lister.Do(context.Background(), &CreatorIDListerDoInput{
		InputCreatorIDs:   []string{"first", "second"},
		IncludeSupporting: true,
		IncludeFollowing:  true,
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"first", "second"}, creatorIDs)
	assert.Empty(t, transport.requestURLs())
}

type staticResponseTransport struct {
	mu        sync.Mutex
	responses map[string]string
	requests  []string
}

func (t *staticResponseTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	url := req.URL.String()
	t.requests = append(t.requests, url)
	body, ok := t.responses[url]
	if !ok {
		return nil, fmt.Errorf("unexpected request URL %q", url)
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     http.StatusText(http.StatusOK),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

func (t *staticResponseTransport) requestURLs() []string {
	t.mu.Lock()
	defer t.mu.Unlock()

	return append([]string(nil), t.requests...)
}

func newStaticOfficialAPIClient(transport http.RoundTripper) *OfficialAPIClient {
	httpClient := retryablehttp.NewClient()
	httpClient.HTTPClient.Transport = transport
	httpClient.RetryMax = 0

	return &OfficialAPIClient{HTTPClient: httpClient}
}
