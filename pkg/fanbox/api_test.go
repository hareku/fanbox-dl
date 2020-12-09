package fanbox

import (
	"reflect"
	"testing"
)

func stringPointer(str string) *string {
	return &str
}

func TestPostBody_OrderedImageMap(t *testing.T) {
	type fields struct {
		Blocks   *[]Block
		Images   *[]Image
		ImageMap *map[string]Image
	}

	tests := []struct {
		name   string
		fields fields
		want   []Image
	}{
		{
			name: "sort images by blocks order",
			fields: fields{
				Images: nil,
				Blocks: &[]Block{
					{
						Type:    "image",
						ImageID: stringPointer("first"),
					},
					{
						Type:    "image",
						ImageID: stringPointer("second"),
					},
				},
				ImageMap: &map[string]Image{
					"second": {
						ID: "second_image",
					},
					"first": {
						ID: "first_image",
					},
				},
			},
			want: []Image{
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
				Images:   &[]Image{},
				Blocks:   nil,
				ImageMap: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &PostBody{
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
