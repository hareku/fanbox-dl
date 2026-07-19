package fanbox

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestPostInfoResponseUnmarshal(t *testing.T) {
	const response = `{
		"body": {
			"post": {
				"id": "12345678",
				"title": "Test_title",
				"feeRequired": 500,
				"isRestricted": false,
				"body": {
					"blocks": [],
					"imageMap": {},
					"fileMap": {}
				}
			}
		}
	}`

	var got PostInfoResponse
	if err := json.Unmarshal([]byte(response), &got); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	post := got.Body.Post
	if post.ID != "12345678" {
		t.Errorf("post ID = %q, want %q", post.ID, "12345678")
	}
	if post.FeeRequired != 500 {
		t.Errorf("fee required = %d, want 500", post.FeeRequired)
	}
	if post.Body == nil {
		t.Fatal("post body is nil")
	}
}
