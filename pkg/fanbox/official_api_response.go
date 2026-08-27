package fanbox

import (
	"bytes"
	"encoding/json"
	"fmt"
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
	Body Post `json:"body"`
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

	if err := json.Unmarshal(postJSON, &r.Body); err != nil {
		return fmt.Errorf("decode post info response post: %w", err)
	}
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
}

type Block struct {
	Type    string  `json:"type"` // p(text) or image.
	ImageID *string `json:"imageId"`
	FileID  *string `json:"fileId"`
}

type Downloadable interface {
	GetID() string
	GetURL() string
	GetThumbnailURL() (string, bool)
	GetExtension() string
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
