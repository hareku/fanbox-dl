package download

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

// Client is the client which downloads images from FANBOX.
type Client struct {
	UserID         string
	SaveDir        string
	FANBOXSESSID   string
	SeparateByPost bool
}

// Run downloads images.
func (c *Client) Run(ctx context.Context) error {
	url := c.buildFirstURL()

	for {
		resp, err := c.request(ctx, url)
		if err != nil {
			return fmt.Errorf("request error (%s): %w", url, err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("body reading error: %w", err)
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("status code is %d, response body: %s", resp.StatusCode, body)
		}

		var content ListCreator
		err = json.Unmarshal(body, &content)
		if err != nil {
			return fmt.Errorf("json unmarshal error: %w", err)
		}

		for _, post := range content.Body.Items {
			if post.Body == nil {
				log.Printf("Skipping an unauthorized post: %s", post.Title)
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

// request sends GET request with credentials.
func (c *Client) request(ctx context.Context, url string) (*http.Response, error) {
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("http request building error: %w", err)
	}

	req.Header.Set("Cookie", fmt.Sprintf("FANBOXSESSID=%s", c.FANBOXSESSID))
	req.Header.Set("Origin", "https://www.fanbox.cc")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request error: %w", err)
	}

	return resp, nil
}

func (c *Client) downloadWithRetry(ctx context.Context, post Post, order int, img Image) error {
	const maxRetry = 5
	retry := 0
	var err error

	for {
		if retry >= maxRetry {
			break
		}

		err = c.download(ctx, post, order, img)
		if err == nil {
			break
		}

		// Not retryable error
		if !errors.Is(err, io.ErrUnexpectedEOF) {
			break
		}

		time.Sleep(time.Second)
		retry++
	}

	if err != nil {
		return fmt.Errorf("failed to download with retry %d times: %w", retry, err)
	}

	return nil
}

func (c *Client) makeFileName(post Post, order int, img Image) string {
	date, err := time.Parse(time.RFC3339, post.PublishedDateTime)
	if err != nil {
		panic(fmt.Errorf("failed to parse post published date time %s: %w", post.PublishedDateTime, err))
	}

	if c.SeparateByPost {
		// [SaveDirectory]/2006-01-02-[Post Title]/[Order]-[Image ID].[Image Extension]
		return filepath.Join(c.SaveDir, fmt.Sprintf("%s-%s", date.UTC().Format("2006-01-02"), post.Title), fmt.Sprintf("%d-%s.%s", order, img.ID, img.Extension))
	}

	// [SaveDirectory]/2006-01-02-[Post Title]-[Order]-[Image ID].[Image Extension]
	return filepath.Join(c.SaveDir, fmt.Sprintf("%s-%s-%d-%s.%s", date.UTC().Format("2006-01-02"), post.Title, order, img.ID, img.Extension))
}

func (c *Client) download(ctx context.Context, post Post, order int, img Image) error {
	name := c.makeFileName(post, order, img)

	if c.isDownloaded(name) {
		log.Printf("Already downloaded %dth file of %s\n", order, post.Title)
		return nil
	}

	log.Printf("Downloading %dth file of %s\n", order, post.Title)

	resp, err := c.request(ctx, img.OriginalURL)
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

func (c *Client) saveFile(name string, resp *http.Response) error {
	dir := filepath.Dir(name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0775)
		if err != nil {
			return fmt.Errorf("failed to create a directory (%s): %w", dir, err)
		}
	}

	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0775)
	if err != nil {
		return fmt.Errorf("failed to open a file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		// Remove crashed file
		errRemove := os.Remove(file.Name())
		if errRemove != nil {
			return fmt.Errorf("file copying error and couldn't remove a crashed file (%s): %w", file.Name(), errRemove)
		}

		return fmt.Errorf("file copying error: %w", err)
	}

	return nil
}

func (c *Client) isDownloaded(name string) bool {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}

	return true
}
