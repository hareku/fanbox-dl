package fanbox

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// Pagination represents the response of https://api.fanbox.cc/post.paginateCreator?creatorId=x.
type Pagination struct {
	Pages []string `json:"body"`
}

// UnmarshalJSON supports both the legacy body array and the current body.pageUrls array.
func (p *Pagination) UnmarshalJSON(data []byte) error {
	pages, err := decodeResponseBodyArray[string](data, "pageUrls")
	if err != nil {
		return fmt.Errorf("decode pagination response: %w", err)
	}

	p.Pages = pages
	return nil
}

// ListCreatorResponse represents the response of https://api.fanbox.cc/post.listCreator.
type ListCreatorResponse struct {
	Body []Post `json:"body"`
}

// UnmarshalJSON supports both the legacy body array and the current body.posts array.
func (r *ListCreatorResponse) UnmarshalJSON(data []byte) error {
	posts, err := decodeResponseBodyArray[Post](data, "posts")
	if err != nil {
		return fmt.Errorf("decode list creator response: %w", err)
	}

	r.Body = posts
	return nil
}

// PostInfoResponse represents the response of https://api.fanbox.cc/post.info.
type PostInfoResponse struct {
	Body PostInfoBody `json:"body"`
}

type PostInfoBody struct {
	Post      Post   `json:"post"`
	ID        string `json:"id,omitempty"`
	Title     string `json:"title,omitempty"`
	CreatorID string `json:"creatorId,omitempty"`
}

// UnmarshalJSON supports both the legacy body object and the current body.post object.
// A body without a post decodes to a zero Post so callers can skip it.
func (r *PostInfoResponse) UnmarshalJSON(data []byte) error {
	*r = PostInfoResponse{}

	body, err := decodeResponseBody(data)
	if err != nil {
		return fmt.Errorf("decode post info response: %w", err)
	}
	if body == nil {
		return nil
	}

	var probe struct {
		Post json.RawMessage `json:"post"`
		ID   json.RawMessage `json:"id"`
	}
	if err := json.Unmarshal(body, &probe); err != nil {
		return fmt.Errorf("decode post info response body: %w", err)
	}

	postJSON := probe.Post
	if isNullJSON(postJSON) {
		if isNullJSON(probe.ID) {
			return nil
		}
		postJSON = body
	}

	var post Post
	if err := json.Unmarshal(postJSON, &post); err != nil {
		return fmt.Errorf("decode post info response post: %w", err)
	}
	r.Body.Post = post
	r.Body.ID, r.Body.Title, r.Body.CreatorID = post.ID, post.Title, post.CreatorID
	return nil
}

type responseEnvelope struct {
	Body json.RawMessage `json:"body"`
}

func decodeResponseBody(data []byte) (json.RawMessage, error) {
	var response responseEnvelope
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("decode response envelope: %w", err)
	}

	body := bytes.TrimSpace(response.Body)
	if isNullJSON(body) {
		return nil, nil
	}
	return body, nil
}

func decodeResponseBodyArray[T any](data []byte, wrappedField string) ([]T, error) {
	body, err := decodeResponseBody(data)
	if err != nil {
		return nil, err
	}
	if body == nil {
		return nil, nil
	}

	if body[0] == '{' {
		var fields map[string]json.RawMessage
		if err := json.Unmarshal(body, &fields); err != nil {
			return nil, fmt.Errorf("decode response body: %w", err)
		}

		wrapped, ok := fields[wrappedField]
		if !ok {
			return nil, fmt.Errorf("response body does not contain %q", wrappedField)
		}
		if isNullJSON(wrapped) {
			return nil, nil
		}

		var values []T
		if err := json.Unmarshal(wrapped, &values); err != nil {
			return nil, fmt.Errorf("decode response body field %q: %w", wrappedField, err)
		}
		return values, nil
	}

	var values []T
	if err := json.Unmarshal(body, &values); err != nil {
		return nil, fmt.Errorf("decode response body: %w", err)
	}
	return values, nil
}

func isNullJSON(raw json.RawMessage) bool {
	raw = bytes.TrimSpace(raw)
	return len(raw) == 0 || bytes.Equal(raw, []byte("null"))
}

// Post represents post attributes.
type Post struct {
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	PublishedDateTime string    `json:"publishedDatetime"`
	CreatorID         string    `json:"creatorId"`
	FeeRequired       int       `json:"feeRequired"`
	IsRestricted      bool      `json:"isRestricted"`
	IsPinned          bool      `json:"isPinned"`
	Body              *PostBody `json:"body"`
}

type PostBody struct {
	// Files is not nil if post type is "file".
	Files *[]File `json:"files"`
	// Images is not nil if post type is "image".
	Images *[]Image `json:"images"`
	// Blocks is not nil if post type is "blog".
	Blocks *[]Block `json:"blocks"`
	// ImageMap is not nil if post type is "blog".
	ImageMap *map[string]Image `json:"imageMap"`
	// FileMap is not nil if post type is "blog".
	FileMap *map[string]File `json:"fileMap"`
	// Text is for simple text in post type "image"/"file".
	Text string `json:"text"`
}

func (p *Post) GetTextContent() string {
	if p.Body != nil {
		return p.Body.ExtractText()
	}
	return ""
}

func (pb *PostBody) ExtractText() string {
	var textContent strings.Builder

	// Handle article/blog type posts with blocks
	if pb.Blocks != nil {
		for _, block := range *pb.Blocks {
			if block.Type == "p" && block.Text != "" {
				textContent.WriteString(block.Text)
				textContent.WriteString("\n\n")
			} else if block.Type == "image" && block.ImageID != nil {
				textContent.WriteString("[Image: " + *block.ImageID + "]\n\n")
			} else if block.Type == "file" && block.FileID != nil {
				fileInfo := (*pb.FileMap)[*block.FileID]
				fileLabel := fileInfo.ID
				if fileInfo.Name != "" {
					fileLabel = fileInfo.Name
				}
				textContent.WriteString("[File: " + fileLabel + "." + fileInfo.Extension + "]\n\n")
			}
		}
	}

	// Handle image type posts with text field
	if textContent.Len() == 0 && pb.Text != "" {
		textContent.WriteString(pb.Text)

		// Add file information at the end for file-type posts
		if pb.Files != nil {
			textContent.WriteString("\n\n--- Files ---\n")
			for _, file := range *pb.Files {
				fileLabel := file.ID
				if file.Name != "" {
					fileLabel = file.Name
				}
				textContent.WriteString(fmt.Sprintf("[File: %s.%s]\n", fileLabel, file.Extension))
			}
		}

		// Add image information at the end for image-type posts
		if pb.Images != nil {
			textContent.WriteString("\n\n--- Images ---\n")
			for _, img := range *pb.Images {
				textContent.WriteString(fmt.Sprintf("[Image: %s.%s]\n", img.ID, img.Extension))
			}
		}
	}

	return strings.TrimSpace(textContent.String())
}

type Block struct {
	Type    string  `json:"type"` // p(text) or image.
	Text    string  `json:"text"` // Text content for "p" type blocks
	ImageID *string `json:"imageId"`
	FileID  *string `json:"fileId"`
}

type Downloadable interface {
	GetID() string
	GetURL() string
	GetThumbnailURL() (string, bool)
	GetExtension() string
	GetName() string // New method to get the display name
}

// File represents a uploaded file.
type File struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Extension string `json:"extension"`
	URL       string `json:"url"`
}

func (f File) GetID() string {
	return f.ID
}

func (f File) GetURL() string {
	return f.URL
}

func (f File) GetThumbnailURL() (string, bool) {
	return "", false
}

func (f File) GetExtension() string {
	return f.Extension
}

func (f File) GetName() string {
	if f.Name != "" {
		return f.Name
	}
	return f.ID
}

// Image represents a uploaded image.
type Image struct {
	ID           string `json:"id"`
	Extension    string `json:"extension"`
	OriginalURL  string `json:"originalUrl"`
	ThumbnailURL string `json:"thumbnailUrl"`
}

func (i Image) GetID() string {
	return i.ID
}

func (i Image) GetURL() string {
	return i.OriginalURL
}

func (i Image) GetThumbnailURL() (string, bool) {
	return i.ThumbnailURL, true
}

func (i Image) GetExtension() string {
	return i.Extension
}

func (i Image) GetName() string {
	return i.ID // Images don't have names in the API response, use ID instead
}

func (f *Post) ListDownloadable() []Downloadable {
	if f.Body == nil {
		return nil
	}

	if f.Body.Images != nil {
		res := make([]Downloadable, 0, len(*f.Body.Images))
		for _, v := range *f.Body.Images {
			res = append(res, v)
		}
		return res
	}

	if f.Body.Files != nil {
		res := make([]Downloadable, 0, len(*f.Body.Files))
		for _, v := range *f.Body.Files {
			res = append(res, v)
		}
		return res
	}

	if f.Body.Blocks != nil {
		res := make([]Downloadable, 0)
		for _, v := range *f.Body.Blocks {
			if v.ImageID != nil {
				res = append(res, (*f.Body.ImageMap)[*v.ImageID])
			}
			if v.FileID != nil {
				res = append(res, (*f.Body.FileMap)[*v.FileID])
			}
		}
		return res
	}

	return nil
}

type PlanListSupportingResponse struct {
	Body []Plan `json:"body"`
}

// UnmarshalJSON supports both the legacy body array and the current body.plans array.
func (r *PlanListSupportingResponse) UnmarshalJSON(data []byte) error {
	plans, err := decodeResponseBodyArray[Plan](data, "plans")
	if err != nil {
		return fmt.Errorf("decode list supporting plans response: %w", err)
	}

	r.Body = plans
	return nil
}

type PlanListFollowingResponse struct {
	Body PlanListCreators `json:"body"`
}

// UnmarshalJSON supports both the legacy body array and the current body.creators array.
func (r *PlanListFollowingResponse) UnmarshalJSON(data []byte) error {
	creators, err := decodeResponseBodyArray[Plan](data, "creators")
	if err != nil {
		return fmt.Errorf("decode list following creators response: %w", err)
	}

	r.Body.Creators = creators
	return nil
}

type PlanListCreators struct {
	Creators []Plan `json:"creators"`
}

type Plan struct {
	CreatorID string `json:"creatorId"`
}

type CreatorListFollowingResponse struct {
	Body []Creator `json:"body"`
}

type Creator struct {
	CreatorID string `json:"creatorId"`
}
