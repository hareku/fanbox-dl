package fanbox

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	backoff "github.com/cenkalti/backoff/v4"
)

// Client is the struct for Client.
type Client struct {
	CheckAllPosts     bool
	DryRun            bool
	DownloadFiles     bool
	OfficialAPIClient *OfficialAPIClient
	Storage           *LocalStorage
	Logger            *Logger
}

func (c *Client) Run(ctx context.Context, creatorID string) error {
	nextURL := func() string {
		params := url.Values{}
		params.Set("creatorId", creatorID)
		params.Set("limit", "50")

		return fmt.Sprintf("https://api.fanbox.cc/post.listCreator?%s", params.Encode())
	}()

	for {
		content := ListCreatorResponse{}
		err := c.OfficialAPIClient.RequestAndUnwrapJSON(ctx, http.MethodGet, nextURL, &content)
		if err != nil {
			return fmt.Errorf("failed to list posts of %q: %w", creatorID, err)
		}

		for _, item := range content.Body.Items {
			if item.IsRestricted {
				c.Logger.Debugf("Skipping a ristricted post: %q.", item.Title)
				continue
			}

			postResp := PostInfoResponse{}
			err := c.OfficialAPIClient.RequestAndUnwrapJSON(
				ctx, http.MethodGet,
				fmt.Sprintf("https://api.fanbox.cc/post.info?postId=%s", item.ID),
				&postResp)
			if err != nil {
				return fmt.Errorf("failed to get post: %w", err)
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

				isDownloaded, err := c.Storage.Exist(post, order, d)
				if err != nil {
					return fmt.Errorf("failed to check whether does %s exist: %w", assetType, err)
				}

				if isDownloaded {
					c.Logger.Debugf("Already downloaded %dth %s of %q.", order, assetType, post.Title)
					if !c.CheckAllPosts {
						c.Logger.Debugf("No more new files and images.")
						return nil
					}
					continue
				}

				if c.DryRun {
					c.Logger.Infof("[dry-run] Client will download %dth %s of %q.\n", order, assetType, post.Title)
					continue
				}

				c.Logger.Infof("Downloading %dth %s of %s\n", order, assetType, post.Title)
				err = c.downloadImage(ctx, post, order, d)
				if err != nil {
					return fmt.Errorf("download error: %w", err)
				}
			}
		}

		if content.Body.NextURL == nil {
			break
		}
		nextURL = *content.Body.NextURL
	}

	return nil
}

// downloadImage downloads and save the file with retrying.
func (c *Client) downloadImage(ctx context.Context, post Post, order int, d Downloadable) error {
	strategy := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5)
	strategy = backoff.WithContext(strategy, ctx)

	err := backoff.Retry(func() error {
		resp, err := c.OfficialAPIClient.Request(ctx, http.MethodGet, d.GetURL())
		if err != nil {
			return fmt.Errorf("request error (%s): %w", d.GetURL(), err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return fmt.Errorf("status code %d", resp.StatusCode)
		}

		if err := c.Storage.Save(post, order, d, resp.Body); err != nil {
			return fmt.Errorf("failed to save a file: %w", err)
		}

		return nil
	}, strategy)
	if err != nil {
		return fmt.Errorf("failed to request file with retrying: %w", err)
	}

	return nil
}
