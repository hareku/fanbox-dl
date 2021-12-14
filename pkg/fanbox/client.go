package fanbox

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	backoff "github.com/cenkalti/backoff/v4"
)

// Client is the client which downloads images from FANBOX.
type Client interface {
	Run(ctx context.Context, creatorID string) error
}

// client is the struct for Client.
type client struct {
	checkAllPosts bool
	dryRun        bool
	downloadFiles bool
	api           API
	storage       Storage
	fileStorage   FileStorage
}

// NewClientInput is the input of NewClient.
type NewClientInput struct {
	CheckAllPosts bool
	DryRun        bool
	DownloadFiles bool

	API         API
	Storage     Storage
	FileStorage FileStorage
}

// NewClient return the new Client instance.
func NewClient(input *NewClientInput) Client {
	return &client{
		checkAllPosts: input.CheckAllPosts,
		dryRun:        input.DryRun,
		downloadFiles: input.DownloadFiles,
		api:           input.API,
		storage:       input.Storage,
		fileStorage:   input.FileStorage,
	}
}

// Run downloads images.
func (c *client) Run(ctx context.Context, creatorID string) error {
	url := ListCreatorURL(creatorID, 50)

	for {
		content, err := c.api.ListCreator(ctx, url)
		if err != nil {
			return fmt.Errorf("failed to list images of %q: %w", creatorID, err)
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
					return fmt.Errorf("failed to check whether does image exist: %w", err)
				}

				if isDownloaded {
					log.Printf("Already downloaded %dth image of %q.\n", order, post.Title)
					if !c.checkAllPosts {
						log.Println("No more new images.")
						return nil
					}
					continue
				}

				if c.dryRun {
					log.Printf("[dry-run] Client will download %dth image of %q.\n", order, post.Title)
					continue
				}

				log.Printf("Downloading %dth image of %s\n", order, post.Title)
				err = c.downloadImage(ctx, post, order, img)
				if err != nil {
					return fmt.Errorf("download error: %w", err)
				}
			}

			if post.Body.Files != nil {
				for order, f := range *post.Body.Files {
					isDownloaded, err := c.fileStorage.Exist(post, order, f)
					if err != nil {
						return fmt.Errorf("failed to check whether does file exist: %w", err)
					}

					if isDownloaded {
						log.Printf("Already downloaded %dth file of %q.\n", order, post.Title)
						if !c.checkAllPosts {
							log.Println("No more new files.")
							return nil
						}
						continue
					}

					if c.dryRun {
						log.Printf("[dry-run] Client will download %dth file of %q.\n", order, post.Title)
						continue
					}

					log.Printf("Downloading %dth file of %s\n", order, post.Title)
					err = c.downloadFile(ctx, post, order, f)
					if err != nil {
						return fmt.Errorf("download error: %w", err)
					}
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

// downloadImage downloads and save the image with retrying.
func (c *client) downloadImage(ctx context.Context, post Post, order int, img Image) error {
	strategy := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5)
	strategy = backoff.WithContext(strategy, ctx)

	var rc io.ReadCloser
	err := backoff.Retry(func() error {
		resp, err := c.api.Request(ctx, http.MethodGet, img.OriginalURL)
		if err != nil {
			return fmt.Errorf("request error (%s): %w", img.OriginalURL, err)
		}

		if resp.StatusCode != 200 {
			defer resp.Body.Close()
			return fmt.Errorf("status code %d", resp.StatusCode)
		}

		rc = resp.Body
		return nil
	}, strategy)
	if err != nil {
		return fmt.Errorf("failed to request file with retrying: %w", err)
	}
	defer rc.Close()

	if err := c.storage.Save(post, order, img, rc); err != nil {
		return fmt.Errorf("failed to save a file: %w", err)
	}

	return nil
}

// downloadFile downloads and save the image with retrying.
func (c *client) downloadFile(ctx context.Context, post Post, order int, f File) error {
	strategy := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5)
	strategy = backoff.WithContext(strategy, ctx)

	var rc io.ReadCloser
	err := backoff.Retry(func() error {
		resp, err := c.api.Request(ctx, http.MethodGet, f.URL)
		if err != nil {
			return fmt.Errorf("request error (%s): %w", f.URL, err)
		}

		if resp.StatusCode != 200 {
			defer resp.Body.Close()
			return fmt.Errorf("status code %d", resp.StatusCode)
		}

		rc = resp.Body
		return nil
	}, strategy)
	if err != nil {
		return fmt.Errorf("failed to request file with retrying: %w", err)
	}
	defer rc.Close()

	if err := c.fileStorage.Save(post, order, f, rc); err != nil {
		return fmt.Errorf("failed to save a file: %w", err)
	}

	return nil
}
