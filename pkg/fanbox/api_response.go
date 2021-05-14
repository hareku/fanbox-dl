package fanbox

import (
	"fmt"
	"net/url"
	"strconv"
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

// buildListCreatorURL builds the first page URL of /post.listCreator.
func buildListCreatorURL(userID string, perPage int) string {
	params := url.Values{}
	params.Set("creatorId", userID)
	params.Set("limit", strconv.Itoa(perPage))

	return fmt.Sprintf("https://api.fanbox.cc/post.listCreator?%s", params.Encode())
}
