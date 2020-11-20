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
	"time"

	"github.com/hareku/fanbox-dl/pkg/api"
)

// Client is the client which downloads images from FANBOX.
type Client struct {
	UserID         string
	SaveDir        string
	FANBOXSESSID   string
	SeparateByPost bool
	CheckAllPosts  bool
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

		var content api.ListCreator
		err = json.Unmarshal(body, &content)
		if err != nil {
			return fmt.Errorf("json unmarshal error: %w", err)
		}

		for _, post := range content.Body.Items {
			if post.Body == nil {
				log.Printf("Skipping an unauthorized post: %q.\n", post.Title)
				continue
			}

			var images []api.Image
			if post.Body.Images != nil {
				images = *post.Body.Images
			}
			if images == nil && post.Body.ImageMap != nil {
				images = post.Body.OrderedImageMap()
			}

			for order, img := range images {
				if c.isDownloaded(c.makeFileName(post, order, img)) {
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

// request sends GET request with credentials.
func (c *Client) request(ctx context.Context, url string) (*http.Response, error) {
	resp, err := api.Request(ctx, c.FANBOXSESSID, url)
	if err != nil {
		return nil, fmt.Errorf("http request error: %w", err)
	}

	return resp, nil
}

func (c *Client) downloadWithRetry(ctx context.Context, post api.Post, order int, img api.Image) error {
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

		// HTTP body often disconnects and returns error io.ErrUnexpectedEOF.
		// But if err is not io.ErrUnexpectedEOF, stop the retrying.
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

func (c *Client) download(ctx context.Context, post api.Post, order int, img api.Image) error {
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
