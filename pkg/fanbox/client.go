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

	"golang.org/x/net/http2"
)

// Client is the struct for Client.
type Client struct {
	CheckAllPosts     bool
	DryRun            bool
	SkipFiles         bool
	SkipOnError       bool
	OfficialAPIClient *OfficialAPIClient
	Storage           *LocalStorage
}

func (c *Client) Run(ctx context.Context, creatorID string) error {
	var pagination Pagination
	if err := c.OfficialAPIClient.RequestAndUnwrapJSON(
		ctx, http.MethodGet,
		fmt.Sprintf("https://api.fanbox.cc/post.paginateCreator?creatorId=%s", url.QueryEscape(creatorID)),
		&pagination); err != nil {
		return fmt.Errorf("get pagination: %w", err)
	}
	slog.Debug("Found pages", slog.Int("count", len(pagination.Pages)), slog.String("creatorID", creatorID))

	for _, page := range pagination.Pages {
		content := ListCreatorResponse{}
		err := c.OfficialAPIClient.RequestAndUnwrapJSON(ctx, http.MethodGet, page, &content)
		if err != nil {
			return fmt.Errorf("list posts of %q: %w", creatorID, err)
		}
		slog.Debug("Found posts", slog.Int("count", len(content.Body)), slog.String("page", page))

		for _, item := range content.Body {
			if item.IsRestricted {
				slog.Debug("Skipping restricted post", slog.String("publishedDateTime", item.PublishedDateTime), slog.String("title", item.Title))
				continue
			}

			postResp := PostInfoResponse{}
			err := c.OfficialAPIClient.RequestAndUnwrapJSON(
				ctx, http.MethodGet,
				fmt.Sprintf("https://api.fanbox.cc/post.info?postId=%s", item.ID),
				&postResp)
			if err != nil {
				return fmt.Errorf("get post: %w", err)
			}
			post := postResp.Body

			// for backward-compatibility, split downloadable file's order into two.
			imgOrder := 0
			fileOrder := 0

			for _, d := range post.ListDownloadable() {
				var order int
				var assetType string

				switch d.(type) {
				case Image:
					assetType = "image"
					order = imgOrder
					imgOrder++
				case File:
					assetType = "file"
					order = fileOrder
					fileOrder++
				default:
					return fmt.Errorf("unsupported asset type: %+v", d)
				}

				if d.GetID() == "" {
					slog.Info("Can't download", slog.Int("i", order), slog.String("title", post.Title), slog.String("reason", "bad URL"))
					continue
				}

				isDownloaded, err := c.Storage.Exist(post, order, d)
				if err != nil {
					return fmt.Errorf("check whether does %s exist: %w", assetType, err)
				}

				if isDownloaded {
					slog.Debug("Already downloaded", slog.Int("i", order), slog.String("title", post.Title))
					if !c.CheckAllPosts {
						slog.Debug("No more new files and images")
						return nil
					}
					continue
				}

				if assetType == "file" && c.SkipFiles {
					slog.Debug("Skipping file", slog.Int("order", order), slog.String("title", post.Title))
					continue
				}

				if c.DryRun {
					slog.Info("[dry-run] Client will download", slog.Int("order", order), slog.String("assetType", assetType), slog.String("title", post.Title))
					continue
				}

				slog.Info("Downloading", slog.Int("order", order), slog.String("assetType", assetType), slog.String("title", post.Title))
				if err := c.downloadWithRetry(ctx, post, order, d); err != nil {
					if c.SkipOnError {
						slog.Error("Skip downloading due to error", slog.String("error", err.Error()))
						continue
					}
					return fmt.Errorf("download: %w", err)
				}
			}
		}
	}

	return nil
}

func (c *Client) downloadWithRetry(ctx context.Context, post Post, order int, d Downloadable) error {
	for retry := 0; retry < 10; retry++ {
		if retry > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(time.Second):
			}
		}

		if err := c.download(ctx, post, order, d); err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				slog.Error("Download error, retrying", slog.String("type", "io.ErrUnexpectedEOF"), slog.String("error", err.Error()))
				continue
			}

			// fanbox API sometimes forcibly closes the connection when downloading files many times, so retry.
			var opErr *net.OpError
			if errors.As(err, &opErr) {
				slog.Error("Download error, retrying", slog.String("type", "net.OpError"), slog.String("error", opErr.Error()))
				continue
			}

			var goAwayErr *http2.GoAwayError
			if errors.As(err, &goAwayErr) {
				slog.Error("Download error, retrying", slog.String("type", "http2.GoAwayError"), slog.String("error", goAwayErr.Error()))
				continue
			}

			return fmt.Errorf("download error: %w", err)
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
