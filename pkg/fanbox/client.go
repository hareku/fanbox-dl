package fanbox

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"

	backoff "github.com/cenkalti/backoff/v4"
)

// Client is the client which downloads images from
type Client struct {
	UserID         string
	SaveDir        string
	FANBOXSESSID   string
	SeparateByPost bool
	CheckAllPosts  bool
	DryRun         bool
}

// Run downloads images.
func (c *Client) Run(ctx context.Context) error {
	url := c.buildFirstURL()

	for {
		content, err := c.fetchListCreator(ctx, url)
		if err != nil {
			return fmt.Errorf("failed to fetch %q: %w", url, err)
		}

		for _, post := range content.Body.Items {
			if post.Body == nil {
				log.Printf("Skipping an unauthorized post: %q.\n", post.Title)
				continue
			}

			var images []Image
			if post.Body.Images != nil {
				images = *post.Body.Images
			}
			if images == nil && post.Body.ImageMap != nil {
				images = post.Body.OrderedImageMap()
			}

			for order, img := range images {
				downloaded, err := c.isDownloaded(c.makeFileName(post, order, img))
				if err != nil {
					return fmt.Errorf("failed to check whether does file exist: %w", err)
				}

				if downloaded {
					log.Printf("Already downloaded %dth file of %q.\n", order, post.Title)
					if !c.CheckAllPosts {
						log.Println("No more new images.")
						return nil
					}
				}

				err = c.downloadWithRetry(ctx, post, order, img)
				if err != nil {
					return fmt.Errorf("download error: %w", err)
				}
			}
		}

		if content.Body.NextURL == nil {
			break
		}

		url = *content.Body.NextURL
	}

	return nil
}

func (c *Client) buildFirstURL() string {
	params := url.Values{}
	params.Set("creatorId", c.UserID)
	params.Set("limit", "50")

	return fmt.Sprintf("https://api.fanbox.cc/post.listCreator?%s", params.Encode())
}

// fetchListCreator fetches the ListCreator sturct by URL.
func (c *Client) fetchListCreator(ctx context.Context, url string) (*ListCreator, error) {
	var list ListCreator

	operation := func() error {
		err := RequestAsJSON(ctx, c.FANBOXSESSID, url, &list)
		if err != nil {
			return fmt.Errorf("failed to fetch ListCreator: %w", err)
		}

		return nil
	}

	strategy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)

	err := backoff.Retry(operation, strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch ListCreator retry: %w", err)
	}

	return &list, nil
}

func (c *Client) downloadWithRetry(ctx context.Context, post Post, order int, img Image) error {
	operation := func() error {
		err := c.download(ctx, post, order, img)
		if err == nil {
			return nil
		}

		// HTTP body often disconnects while downloading and returns error io.ErrUnexpectedEOF.
		// But other errors may be permanent error, so return and wrap it.
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return err
		}

		return backoff.Permanent(err)
	}

	strategy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)

	err := backoff.Retry(operation, strategy)
	if err != nil {
		return fmt.Errorf("failed to download with retry: %w", err)
	}

	return nil
}

func (c *Client) download(ctx context.Context, post Post, order int, img Image) error {
	name := c.makeFileName(post, order, img)
	if c.DryRun {
		log.Printf("[dry-run] Client will download %dth file of %q.\n", order, post.Title)
		return nil
	}

	log.Printf("Downloading %dth file of %s\n", order, post.Title)

	resp, err := Request(ctx, c.FANBOXSESSID, img.OriginalURL)
	if err != nil {
		return fmt.Errorf("request error (%s): %w", img.OriginalURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("file (%s) status code is %d", img.OriginalURL, resp.StatusCode)
	}

	err = c.saveFile(name, resp)
	if err != nil {
		return fmt.Errorf("failed to save an image: %w", err)
	}

	return nil
}
