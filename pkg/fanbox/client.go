package fanbox

import (
	"context"
	"fmt"
	"log"
	"net/http"

	backoff "github.com/cenkalti/backoff/v4"
)

// Client is the client which downloads images from FANBOX.
type Client interface {
	Run(ctx context.Context) error
}

// client is the struct for Client.
type client struct {
	userID        string
	checkAllPosts bool
	dryRun        bool
	api           API
	storage       Storage
}

// NewClientInput is the input of NewClient.
type NewClientInput struct {
	UserID        string
	CheckAllPosts bool
	DryRun        bool

	API     API
	Storage Storage
}

// NewClient return the new Client instance.
func NewClient(input *NewClientInput) Client {
	return &client{
		userID:        input.UserID,
		checkAllPosts: input.CheckAllPosts,
		dryRun:        input.DryRun,
		api:           input.API,
		storage:       input.Storage,
	}
}

// Run downloads images.
func (c *client) Run(ctx context.Context) error {
	url := ListCreatorURL(c.userID, 50)

	for {
		content, err := c.api.ListCreator(ctx, url)
		if err != nil {
			return fmt.Errorf("failed to list images of %q: %w", c.userID, err)
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
				isDownloaded, err := c.storage.Exist(post, order, img)
				if err != nil {
					return fmt.Errorf("failed to check whether does file exist: %w", err)
				}

				if isDownloaded {
					log.Printf("Already downloaded %dth file of %q.\n", order, post.Title)
					if !c.checkAllPosts {
						log.Println("No more new images.")
						return nil
					}
					continue
				}

				if c.dryRun {
					log.Printf("[dry-run] Client will download %dth file of %q.\n", order, post.Title)
					continue
				}

				log.Printf("Downloading %dth file of %s\n", order, post.Title)
				err = c.downloadImageWithRetrying(ctx, post, order, img)
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

// downloadImageWithRetrying downloads and save the image with retrying.
func (c *client) downloadImageWithRetrying(ctx context.Context, post Post, order int, img Image) error {
	operation := func() error {
		return c.downloadImage(ctx, post, order, img)
	}

	strategy := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5)
	strategy = backoff.WithContext(strategy, ctx)

	err := backoff.Retry(operation, strategy)
	if err != nil {
		return fmt.Errorf("failed to download with retrying: %w", err)
	}

	return nil
}

// downloadImage downloads and save the image.
func (c *client) downloadImage(ctx context.Context, post Post, order int, img Image) error {
	resp, err := c.api.Request(ctx, http.MethodGet, img.OriginalURL)
	if err != nil {
		return fmt.Errorf("request error (%s): %w", img.OriginalURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("file (%s) returns status code %d", img.OriginalURL, resp.StatusCode)
	}

	err = c.storage.Save(post, order, img, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save an image: %w", err)
	}

	return nil
}
