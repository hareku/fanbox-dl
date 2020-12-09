package fanbox

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
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

// Request sends a request to the specified FANBOX URL with credentials.
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

// RequestAsJSON sends a request to the specified FANBOX URL with credentials,
// and unmarshal the response body as the passed struct.
func RequestAsJSON(ctx context.Context, sessid string, url string, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("v of RequestAsJSON should be a pointer")
	}

	resp, err := Request(ctx, sessid, url)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("body reading error: %w", err)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("status code is %d, response body: %s", resp.StatusCode, body)
	}

	err = json.Unmarshal(body, v)
	if err != nil {
		return fmt.Errorf("json unmarshal error: %w", err)
	}

	return nil
}
