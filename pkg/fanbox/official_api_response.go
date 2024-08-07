package fanbox

// Pagination represents the response of https://api.fanbox.cc/post.paginateCreator?creatorId=x.
type Pagination struct {
	Pages []string `json:"body"`
}

// ListCreatorResponse represents the response of https://api.fanbox.cc/post.listCreator.
type ListCreatorResponse struct {
	Body []Post `json:"body"`
}

// PostInfoResponse represents the response of https://api.fanbox.cc/post.info.
type PostInfoResponse struct {
	Body Post `json:"body"`
}

// Post represents post attributes.
type Post struct {
	ID                string    `json:"id"`
	Title             string    `json:"title"`
	PublishedDateTime string    `json:"publishedDatetime"`
	CreatorID         string    `json:"creatorId"`
	FeeRequired       int       `json:"feeRequired"`
	IsRestricted      bool      `json:"isRestricted"`
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

func (f File) GetExtension() string {
	return f.Extension
}

// Image represents a uploaded image.
type Image struct {
	ID          string `json:"id"`
	Extension   string `json:"extension"`
	OriginalURL string `json:"originalUrl"`
}

func (i Image) GetID() string {
	return i.ID
}

func (i Image) GetURL() string {
	return i.OriginalURL
}

func (i Image) GetExtension() string {
	return i.Extension
}

func (f *Post) ListDownloadable() []Downloadable {
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

type Plan struct {
	CreatorID string `json:"creatorId"`
}

type CreatorListFollowingResponse struct {
	Body []Creator `json:"body"`
}

type Creator struct {
	CreatorID string `json:"creatorId"`
}
