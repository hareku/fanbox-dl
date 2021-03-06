package fanbox

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/hareku/go-filename"
	"github.com/hareku/go-strlimit"
)

// Client is the client which downloads images from FANBOX.
type Client interface {
	Run(ctx context.Context) error
}

// client is the struct for Client.
type client struct {
	userID         string
	saveDir        string
	separateByPost bool
	checkAllPosts  bool
	dryRun         bool
	apiClient      ApiClient
	fileClient     FileClient
}

// NewClientInput is the input of NewClient.
type NewClientInput struct {
	UserID         string
	SaveDir        string
	SeparateByPost bool
	CheckAllPosts  bool
	DryRun         bool

	ApiClient  ApiClient
	FileClient FileClient
}

// NewClient return the new Client instance.
func NewClient(input *NewClientInput) Client {
	return &client{
		userID:         input.UserID,
		saveDir:        input.SaveDir,
		separateByPost: input.SeparateByPost,
		checkAllPosts:  input.CheckAllPosts,
		dryRun:         input.DryRun,
		apiClient:      input.ApiClient,
		fileClient:     input.FileClient,
	}
}

// Run downloads images.
func (c *client) Run(ctx context.Context) error {
	url := buildListCreatorURL(c.userID, 50)

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
				log.Println(c.makeFileName(post, order, img))
				isDownloaded, err := c.fileClient.DoesExist(c.makeFileName(post, order, img))
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

// fetchListCreator fetches ListCreator by URL.
func (c *client) fetchListCreator(ctx context.Context, url string) (*ListCreator, error) {
	var list ListCreator

	operation := func() error {
		err := c.apiClient.RequestAsJSON(ctx, url, &list)
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
	name := c.makeFileName(post, order, img)

	resp, err := c.apiClient.Request(ctx, img.OriginalURL)
	if err != nil {
		return fmt.Errorf("request error (%s): %w", img.OriginalURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("file (%s) returns status code %d", img.OriginalURL, resp.StatusCode)
	}

	err = c.fileClient.Save(name, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save an image: %w", err)
	}

	return nil
}

// limitOsSafely limits the string length for OS safely.
func limitOsSafely(s string) string {
	switch runtime.GOOS {
	case "windows":
		return strlimit.LimitRunesWithEnd(s, 210, "...")
	default:
		return strlimit.LimitBytesWithEnd(s, 250, "...")
	}
}

func (c *client) makeFileName(post Post, order int, img Image) string {
	date, err := time.Parse(time.RFC3339, post.PublishedDateTime)
	if err != nil {
		panic(fmt.Errorf("failed to parse post published date time %s: %w", post.PublishedDateTime, err))
	}

	title := filename.EscapeString(post.Title, "-")

	if c.separateByPost {
		// [SaveDirectory]/[UserID]/2006-01-02-[Post Title]/[Order]-[Image ID].[Image Extension]
		return filepath.Join(c.saveDir, c.userID, limitOsSafely(fmt.Sprintf("%s-%s", date.UTC().Format("2006-01-02"), title)), fmt.Sprintf("%d-%s.%s", order, img.ID, img.Extension))
	}

	// [SaveDirectory]/[UserID]/2006-01-02-[Post Title]-[Order]-[Image ID].[Image Extension]
	return filepath.Join(c.saveDir, c.userID, fmt.Sprintf("%s.%s", limitOsSafely(fmt.Sprintf("%s-%s-%d-%s", date.UTC().Format("2006-01-02"), title, order, img.ID)), img.Extension))
}
