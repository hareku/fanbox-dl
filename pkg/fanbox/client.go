package fanbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/hareku/fanbox-dl/internal/ctxval"
	"golang.org/x/net/http2"
)

// Client is the struct for Client.
type Client struct {
	CheckAllPosts     bool
	DryRun            bool
	SkipFiles         bool
	SkipImages        bool
	SkipOnError       bool
	OfficialAPIClient *OfficialAPIClient
	Storage           *LocalStorage
}

func (c *Client) Run(ctx context.Context, creatorID string) error {
	ctx = ctxval.AddSlogAttrs(ctx, slog.String("creator_id", creatorID))

	var pagination Pagination
	if err := c.OfficialAPIClient.RequestAndUnwrapJSON(
		ctx, http.MethodGet,
		fmt.Sprintf("https://api.fanbox.cc/post.paginateCreator?%s", func() string {
			q := url.Values{}
			q.Set("creatorId", creatorID)
			return q.Encode()
		}()),
		&pagination,
	); err != nil {
		return fmt.Errorf("get pagination: %w", err)
	}
	slog.DebugContext(ctx, "Found pages", "pages", len(pagination.Pages))

	for i, page := range pagination.Pages {
		content := ListCreatorResponse{}
		err := c.OfficialAPIClient.RequestAndUnwrapJSON(ctx, http.MethodGet, page, &content)
		if err != nil {
			return fmt.Errorf("list posts of %q: %w", creatorID, err)
		}
		slog.DebugContext(ctx, "Found posts",
			"page", i+1,
			"posts", len(content.Body),
		)

		if err := c.handlePage(ctx, &content); err != nil {
			if errors.Is(err, errAlreadyDownloaded) {
				slog.DebugContext(ctx, "No more new assets")
				return nil
			}
			if c.SkipOnError {
				slog.ErrorContext(ctx, "Skip downloading page due to error", "error", err)
				continue
			}
			return fmt.Errorf("handle page: %w", err)
		}
	}
	return nil
}

func (c *Client) handlePage(ctx context.Context, content *ListCreatorResponse) error {
	for _, item := range content.Body {
		if err := c.handlePost(ctx, item); err != nil {
			if c.SkipOnError {
				slog.ErrorContext(ctx, "Skip downloading post due to error", "error", err)
				continue
			}
			return fmt.Errorf("handle post: %w", err)
		}
	}
	return nil
}

func (c *Client) handlePost(ctx context.Context, item Post) error {
	ctx = ctxval.AddSlogAttrs(ctx, slog.String("title", item.Title), slog.String("published_at", item.PublishedDateTime))

	if item.IsRestricted {
		slog.DebugContext(ctx, "Skipping restricted post")
		return nil
	}

	postResp := PostInfoResponse{}
	if err := c.OfficialAPIClient.RequestAndUnwrapJSON(
		ctx, http.MethodGet,
		fmt.Sprintf("https://api.fanbox.cc/post.info?%s", func() string {
			q := url.Values{}
			q.Set("postId", item.ID)
			return q.Encode()
		}()),
		&postResp,
	); err != nil {
		return fmt.Errorf("get post: %w", err)
	}
	post := postResp.Body

	// for backward-compatibility, split downloadable file's order into two types
	var (
		nextImgOrder  int
		nextFileOrder int
	)
	for i, d := range post.ListDownloadable() {
		var (
			order     int
			assetType string
		)

		switch d.(type) {
		case Image:
			assetType = "image"
			order = nextImgOrder
			nextImgOrder++
		case File:
			assetType = "file"
			order = nextFileOrder
			nextFileOrder++
		default:
			return fmt.Errorf("unsupported asset type: %+v", d)
		}

		if err := c.handleAsset(
			ctxval.AddSlogAttrs(ctx, slog.Int("i", i), slog.String("asset_type", assetType)),
			post, order, d,
		); err != nil {
			if errors.Is(err, errAlreadyDownloaded) && c.CheckAllPosts {
				continue
			}
			if c.SkipOnError {
				slog.ErrorContext(ctx, "Skip downloading due to error", "error", err)
				continue
			}
			return fmt.Errorf("handle %s: %w", assetType, err)
		}
	}

	return nil
}

var errAlreadyDownloaded = errors.New("already downloaded")

func (c *Client) handleAsset(ctx context.Context, post Post, order int, d Downloadable) error {
	if _, ok := d.(File); ok && c.SkipFiles {
		slog.DebugContext(ctx, "Skip downloading files")
		return nil
	}
	if _, ok := d.(Image); ok && c.SkipImages {
		slog.DebugContext(ctx, "Skip downloading images")
		return nil
	}

	if d.GetID() == "" {
		slog.DebugContext(ctx, "Asset ID is empty")
		return nil
	}

	isDownloaded, err := c.Storage.Exist(post, order, d)
	if err != nil {
		return fmt.Errorf("check whether downloaded: %w", err)
	}

	if isDownloaded {
		slog.DebugContext(ctx, "Already downloaded")
		return errAlreadyDownloaded
	}

	if c.DryRun {
		slog.InfoContext(ctx, "Skip downloading due to dry-run mode")
		return nil
	}

	slog.InfoContext(ctx, "Downloading")
	if err := c.downloadWithRetry(ctx, post, order, d); err != nil {
		return fmt.Errorf("download: %w", err)
	}

	return nil
}

func (c *Client) downloadWithRetry(ctx context.Context, post Post, order int, d Downloadable) error {
	shouldRetry := func(err error) bool {
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return true
		}

		var opErr *net.OpError
		if errors.As(err, &opErr) {
			return true
		}

		var goAwayErr *http2.GoAwayError
		return errors.As(err, &goAwayErr)
	}

	waitDur := time.Second
	for retry := 0; retry < 10; retry++ {
		if err := c.download(ctx, post, order, d); err != nil {
			if !shouldRetry(err) {
				return fmt.Errorf("download error: %w", err)
			}

			slog.ErrorContext(ctx, "Download error, retrying", "error", err, "wait", waitDur)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitDur):
			}
			continue
		}
		break
	}
	return nil
}

func (c *Client) download(ctx context.Context, post Post, order int, d Downloadable) error {
	var resp *http.Response

	resp, err := c.OfficialAPIClient.Request(ctx, http.MethodGet, d.GetURL())
	if err != nil {
		if errors.Is(err, ErrFailedToThumbnailing) {
			slog.InfoContext(ctx, "The original file is not available (maybe it's a too large), so download a thumbnail instead", "original_file", d.GetURL())
			tu, ok := d.GetThumbnailURL()
			if !ok {
				return fmt.Errorf("thumbnail URL is not found")
			}
			slog.InfoContext(ctx, "Downloading a thumbnail", "thumbnail_url", tu)

			resp, err = c.OfficialAPIClient.Request(ctx, http.MethodGet, tu)
			if err != nil {
				return fmt.Errorf("request error (%s): %w", tu, err)
			}
		} else {
			return fmt.Errorf("request error (%s): %w", d.GetURL(), err)
		}
	}

	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.StatusCode != 200 {
		return fmt.Errorf("status code %d", resp.StatusCode)
	}

	if err := c.Storage.Save(post, order, d, resp.Body); err != nil {
		return fmt.Errorf("save a file: %w", err)
	}

	return nil
}
