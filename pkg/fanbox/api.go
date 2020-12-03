package fanbox

import (
	"context"
	"fmt"
	"net/http"
)

// ListCreator represents the response of https://api.fanbox.cc/post.listCreator.
type ListCreator struct {
	Body ListCreatorBody `json:"body"`
}

// ListCreatorBody represents the main content of ListCreator.
type ListCreatorBody struct {
	Items   []Post  `json:"items"`
	NextURL *string `json:"nextUrl"`
}

// Post represents post attributes.
type Post struct {
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	PublishedDateTime string    `json:"publishedDatetime"`
	Body              *PostBody `json:"body"`
}

// PostBody represents a post's body.
// PostBody has "Images" or "Blocks and ImageMap".
type PostBody struct {
	Blocks   *[]Block          `json:"blocks"`
	Images   *[]Image          `json:"images"`
	ImageMap *map[string]Image `json:"imageMap"`
}

// Block represents a text block of a post.
type Block struct {
	Type    string  `json:"type"`
	ImageID *string `json:"imageId"`
}

// Image represents a posted image.
type Image struct {
	ID          string `json:"id"`
	Extension   string `json:"extension"`
	OriginalURL string `json:"originalUrl"`
}

// OrderedImageMap returns ordered images in ImageMap by PostBody.Blocks order.
func (b *PostBody) OrderedImageMap() []Image {
	if b.ImageMap == nil || b.Blocks == nil {
		return nil
	}

	var images []Image

	for _, block := range *b.Blocks {
		if block.Type == "image" && block.ImageID != nil {
			img := (*b.ImageMap)[*block.ImageID]
			images = append(images, img)
		}
	}

	return images
}

// Request sends a request to FANBOX with credentials.
func Request(ctx context.Context, sessid string, url string) (*http.Response, error) {
	client := http.Client{}

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("http request building error: %w", err)
	}

	req.Header.Set("Cookie", fmt.Sprintf("FANBOXSESSID=%s", sessid))
	req.Header.Set("Origin", "https://www.fanbox.cc") // If Origin header is not set, FANBOX returns HTTP 400 error.

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http response error: %w", err)
	}

	return resp, nil
}
