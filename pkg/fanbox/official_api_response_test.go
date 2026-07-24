package fanbox

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	paginationLegacyResponse   = `{"body":["https://api.fanbox.cc/post.listCreator?creatorId=example&limit=10","https://api.fanbox.cc/post.listCreator?creatorId=example&limit=10&page=2"]}`
	paginationCurrentResponse  = `{"body":{"pageUrls":["https://api.fanbox.cc/post.listCreator?creatorId=example&limit=10","https://api.fanbox.cc/post.listCreator?creatorId=example&limit=10&page=2"]}}`
	listCreatorLegacyResponse  = `{"body":[{"id":"1001","title":"First post"},{"id":"1002","title":"Second post"}]}`
	listCreatorCurrentResponse = `{"body":{"posts":[{"id":"1001","title":"First post"},{"id":"1002","title":"Second post"}]}}`
	postInfoLegacyResponse     = `{"body":{"id":"1001","title":"Example post","creatorId":"example"}}`
	postInfoCurrentResponse    = `{"body":{"post":{"id":"1001","title":"Example post","creatorId":"example"}}}`
)

func TestPagination_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{name: "legacy response", data: paginationLegacyResponse},
		{name: "current response", data: paginationCurrentResponse},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response Pagination
			require.NoError(t, json.Unmarshal([]byte(tt.data), &response))
			assert.Equal(t, []string{
				"https://api.fanbox.cc/post.listCreator?creatorId=example&limit=10",
				"https://api.fanbox.cc/post.listCreator?creatorId=example&limit=10&page=2",
			}, response.Pages)
		})
	}
}

func TestListCreatorResponse_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{name: "legacy response", data: listCreatorLegacyResponse},
		{name: "current response", data: listCreatorCurrentResponse},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response ListCreatorResponse
			require.NoError(t, json.Unmarshal([]byte(tt.data), &response))
			require.Len(t, response.Body, 2)
			assert.Equal(t, "1001", response.Body[0].ID)
			assert.Equal(t, "First post", response.Body[0].Title)
			assert.Equal(t, "1002", response.Body[1].ID)
			assert.Equal(t, "Second post", response.Body[1].Title)
		})
	}
}

func TestPostInfoResponse_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		data string
	}{
		{name: "legacy response", data: postInfoLegacyResponse},
		{name: "current response", data: postInfoCurrentResponse},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var response PostInfoResponse
			require.NoError(t, json.Unmarshal([]byte(tt.data), &response))
			assert.Equal(t, "1001", response.Body.ID)
			assert.Equal(t, "Example post", response.Body.Title)
			assert.Equal(t, "example", response.Body.CreatorID)
		})
	}
}

func TestOfficialAPIResponses_UnmarshalJSON_EmptyBody(t *testing.T) {
	tests := []struct {
		name     string
		data     string
		response any
		want     any
	}{
		{
			name:     "missing body",
			data:     `{}`,
			response: &Pagination{},
			want:     &Pagination{},
		},
		{
			name:     "null body",
			data:     `{"body":null}`,
			response: &Pagination{},
			want:     &Pagination{},
		},
		{
			name:     "null wrapped page URLs",
			data:     `{"body":{"pageUrls":null}}`,
			response: &Pagination{},
			want:     &Pagination{},
		},
		{
			name:     "null posts body",
			data:     `{"body":null}`,
			response: &ListCreatorResponse{},
			want:     &ListCreatorResponse{},
		},
		{
			name:     "post info null body",
			data:     `{"body":null}`,
			response: &PostInfoResponse{},
			want:     &PostInfoResponse{},
		},
		{
			name:     "post info empty object body",
			data:     `{"body":{}}`,
			response: &PostInfoResponse{},
			want:     &PostInfoResponse{},
		},
		{
			name:     "post info null wrapped post",
			data:     `{"body":{"post":null}}`,
			response: &PostInfoResponse{},
			want:     &PostInfoResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, json.Unmarshal([]byte(tt.data), tt.response))
			assert.Equal(t, tt.want, tt.response)
		})
	}
}

func TestOfficialAPIResponses_UnmarshalJSON_InvalidBody(t *testing.T) {
	tests := []struct {
		name       string
		data       string
		response   any
		errorMatch string
	}{
		{
			name:       "pagination missing page URLs",
			data:       `{"body":{"unexpected":[]}}`,
			response:   &Pagination{},
			errorMatch: `response body does not contain "pageUrls"`,
		},
		{
			name:       "wrapped posts is not an array",
			data:       `{"body":{"posts":{}}}`,
			response:   &ListCreatorResponse{},
			errorMatch: `decode response body field "posts"`,
		},
		{
			name:       "legacy body is not an array",
			data:       `{"body":"maintenance"}`,
			response:   &Pagination{},
			errorMatch: `decode response body`,
		},
		{
			name:       "post info body is not an object",
			data:       `{"body":"maintenance"}`,
			response:   &PostInfoResponse{},
			errorMatch: `decode post info response body`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := json.Unmarshal([]byte(tt.data), tt.response)
			require.ErrorContains(t, err, tt.errorMatch)
		})
	}
}

func TestPost_ListDownloadable(t *testing.T) {
	ptr := func(s string) *string { return &s }

	tests := []struct {
		name string
		post Post
		want []Downloadable
	}{
		{
			name: "nil body returns nil",
			post: Post{Body: nil},
			want: nil,
		},
		{
			name: "images",
			post: Post{Body: &PostBody{
				Images: &[]Image{
					{ID: "img1", Extension: "jpeg", OriginalURL: "https://example.com/1.jpeg"},
					{ID: "img2", Extension: "png", OriginalURL: "https://example.com/2.png"},
				},
			}},
			want: []Downloadable{
				Image{ID: "img1", Extension: "jpeg", OriginalURL: "https://example.com/1.jpeg"},
				Image{ID: "img2", Extension: "png", OriginalURL: "https://example.com/2.png"},
			},
		},
		{
			name: "files",
			post: Post{Body: &PostBody{
				Files: &[]File{
					{ID: "file1", Name: "a", Extension: "zip", URL: "https://example.com/a.zip"},
				},
			}},
			want: []Downloadable{
				File{ID: "file1", Name: "a", Extension: "zip", URL: "https://example.com/a.zip"},
			},
		},
		{
			name: "blocks with image and file maps",
			post: Post{Body: &PostBody{
				Blocks: &[]Block{
					{Type: "p"},
					{Type: "image", ImageID: ptr("img1")},
					{Type: "file", FileID: ptr("file1")},
				},
				ImageMap: &map[string]Image{
					"img1": {ID: "img1", Extension: "jpeg", OriginalURL: "https://example.com/1.jpeg"},
				},
				FileMap: &map[string]File{
					"file1": {ID: "file1", Name: "a", Extension: "zip", URL: "https://example.com/a.zip"},
				},
			}},
			want: []Downloadable{
				Image{ID: "img1", Extension: "jpeg", OriginalURL: "https://example.com/1.jpeg"},
				File{ID: "file1", Name: "a", Extension: "zip", URL: "https://example.com/a.zip"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.post.ListDownloadable())
		})
	}
}
