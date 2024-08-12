package fanbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"
)

// Client is the struct for Client.
type Client struct {
	CheckAllPosts     bool
	DryRun            bool
	SkipFiles         bool
	OfficialAPIClient *OfficialAPIClient
	Storage           *LocalStorage
	Logger            *Logger
}

func (c *Client) Run(ctx context.Context, creatorID string) error {
	var pagination Pagination
	if err := c.OfficialAPIClient.RequestAndUnwrapJSON(
		ctx, http.MethodGet,
		fmt.Sprintf("https://api.fanbox.cc/post.paginateCreator?creatorId=%s", url.QueryEscape(creatorID)),
		&pagination); err != nil {
		return fmt.Errorf("get pagination: %w", err)
	}
	c.Logger.Debugf("Found %d pages for %s", len(pagination.Pages), creatorID)

	for _, page := range pagination.Pages {
		content := ListCreatorResponse{}
		err := c.OfficialAPIClient.RequestAndUnwrapJSON(ctx, http.MethodGet, page, &content)
		if err != nil {
			return fmt.Errorf("list posts of %q: %w", creatorID, err)
		}
		c.Logger.Debugf("Found %d posts in %s", len(content.Body), page)

		for _, item := range content.Body {
			if item.IsRestricted {
				c.Logger.Debugf("Skipping a ristricted post: %s %q.", item.PublishedDateTime, item.Title)
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
					c.Logger.Infof("Can't download %dth %s of %q: bad URL", order, assetType, post.Title)
					continue
				}

				isDownloaded, err := c.Storage.Exist(post, order, d)
				if err != nil {
					return fmt.Errorf("check whether does %s exist: %w", assetType, err)
				}

				if isDownloaded {
					c.Logger.Debugf("Already downloaded %dth %s of %q.", order, assetType, post.Title)
					if !c.CheckAllPosts {
						c.Logger.Debugf("No more new files and images.")
						return nil
					}
					continue
				}

				if assetType == "file" && c.SkipFiles {
					c.Logger.Debugf("Skipping %dth file (not images) of %q.\n", order, post.Title)
					continue
				}

				if c.DryRun {
					c.Logger.Infof("[dry-run] Client will download %dth %s of %q.\n", order, assetType, post.Title)
					continue
				}

				c.Logger.Infof("Downloading %dth %s of %s\n", order, assetType, post.Title)
				if err := c.downloadWithRetry(ctx, post, order, d); err != nil {
					return fmt.Errorf("download with retry: %w", err)
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
				c.Logger.Errorf("Download error(io.ErrUnexpectedEOF), retrying. %s", err.Error())
				continue
			}

			// fanbox API sometimes forcibly closes the connection when downloading files many times, so retry.
			var opErr *net.OpError
			if errors.As(err, &opErr) {
				c.Logger.Errorf("Download error(net.OpError), retrying. %s", opErr.Error())
				continue
			}

			return fmt.Errorf("download error: %w", err)
		}
		break
	}

	return nil
}

func (c *Client) download(ctx context.Context, post Post, order int, d Downloadable) error {
	resp, err := c.OfficialAPIClient.Request(ctx, http.MethodGet, d.GetURL())
	if err != nil {
		return fmt.Errorf("request error (%s): %w", d.GetURL(), err)
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
