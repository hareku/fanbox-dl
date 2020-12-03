package fanbox_test

import (
	"reflect"
	"testing"

	"github.com/hareku/fanbox-dl/pkg/fanbox"
)

func stringPointer(str string) *string {
	return &str
}

func TestPostBody_OrderedImageMap(t *testing.T) {
	type fields struct {
		Blocks   *[]fanbox.Block
		Images   *[]fanbox.Image
		ImageMap *map[string]fanbox.Image
	}

	tests := []struct {
		name   string
		fields fields
		want   []fanbox.Image
	}{
		{
			name: "sort images by blocks order",
			fields: fields{
				Images: nil,
				Blocks: &[]fanbox.Block{
					{
						Type:    "image",
						ImageID: stringPointer("first"),
					},
					{
						Type:    "image",
						ImageID: stringPointer("second"),
					},
				},
				ImageMap: &map[string]fanbox.Image{
					"second": {
						ID: "second_image",
					},
					"first": {
						ID: "first_image",
					},
				},
			},
			want: []fanbox.Image{
				{
					ID: "first_image",
				},
				{
					ID: "second_image",
				},
			},
		},
		{
			name: "return nil if Blocks and ImageMap are nil",
			fields: fields{
				Images:   &[]fanbox.Image{},
				Blocks:   nil,
				ImageMap: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &fanbox.PostBody{
				Blocks:   tt.fields.Blocks,
				Images:   tt.fields.Images,
				ImageMap: tt.fields.ImageMap,
			}
			if got := b.OrderedImageMap(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PostBody.OrderedImageMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
