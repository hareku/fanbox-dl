package fanbox

import (
	"context"
	"fmt"
	"log"
	"net/url"

	backoff "github.com/cenkalti/backoff/v4"
)

// Client is the client which downloads images from FANBOX.
type Client interface {
	Run(ctx context.Context) error
}

// client is the struct for Client.
type client struct {
	userID         string
	saveDir        string
	sessionID      string
	separateByPost bool
	checkAllPosts  bool
	dryRun         bool
}

// NewClientInput is the input of NewClient.
type NewClientInput struct {
	UserID         string
	SaveDir        string
	FANBOXSESSID   string
	SeparateByPost bool
	CheckAllPosts  bool
	DryRun         bool
}

// NewClient return the new Client instance.
func NewClient(input *NewClientInput) Client {
	return &client{
		userID:         input.UserID,
		saveDir:        input.SaveDir,
		sessionID:      input.FANBOXSESSID,
		separateByPost: input.SeparateByPost,
		checkAllPosts:  input.CheckAllPosts,
		dryRun:         input.DryRun,
	}
}

// Run downloads images.
func (c *client) Run(ctx context.Context) error {
	url := c.buildFirstURL()

	for {
		content, err := c.fetchListCreator(ctx, url)
		if err != nil {
			return fmt.Errorf("failed to list images of %q: %w", url, err)
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
					if !c.checkAllPosts {
						log.Println("No more new images.")
						return nil
					}
					continue
				}

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

// buildFirstURL builds the first page URL of /post.listCreator.
func (c *client) buildFirstURL() string {
	params := url.Values{}
	params.Set("creatorId", c.userID)
	params.Set("limit", "50")

	return fmt.Sprintf("https://api.fanbox.cc/post.listCreator?%s", params.Encode())
}

// fetchListCreator fetches ListCreator by URL.
func (c *client) fetchListCreator(ctx context.Context, url string) (*ListCreator, error) {
	var list ListCreator

	operation := func() error {
		err := RequestAsJSON(ctx, c.sessionID, url, &list)
		if err != nil {
			return fmt.Errorf("failed to request ListCreator: %w", err)
		}

		return nil
	}
	strategy := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)

	err := backoff.Retry(operation, strategy)
	if err != nil {
		return nil, fmt.Errorf("failed to request ListCreator with retrying: %w", err)
	}

	return &list, nil
}

func (c *client) downloadImageWithRetrying(ctx context.Context, post Post, order int, img Image) error {
	operation := func() error {
		return c.downloadImage(ctx, post, order, img)
	}

	strategy := backoff.WithContext(backoff.WithMaxRetries(backoff.NewExponentialBackOff(), 5), ctx)

	err := backoff.Retry(operation, strategy)
	if err != nil {
		return fmt.Errorf("failed to download with retrying: %w", err)
	}

	return nil
}

func (c *client) downloadImage(ctx context.Context, post Post, order int, img Image) error {
	name := c.makeFileName(post, order, img)
	if c.dryRun {
		log.Printf("[dry-run] Client will download %dth file of %q.\n", order, post.Title)
		return nil
	}

	log.Printf("Downloading %dth file of %s\n", order, post.Title)

	resp, err := Request(ctx, c.sessionID, img.OriginalURL)
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
