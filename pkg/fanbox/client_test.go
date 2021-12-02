package fanbox

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"

	"github.com/golang/mock/gomock"
)

var TempDir string

func TestMain(m *testing.M) {
	tempDir, err := ioutil.TempDir(os.TempDir(), "*")
	if err != nil {
		log.Panicf("failed to make tempDir: %s", err.Error())
		return
	}
	TempDir = tempDir

	status := m.Run()

	err = os.RemoveAll(tempDir)
	if err != nil {
		log.Panicf("failed to delete tempDir: %s", err.Error())
		return
	}
	tempDir = ""

	os.Exit(status)
}

func Test_client_Run(t *testing.T) {
	type clientFactory func(t *testing.T) client
	tests := []struct {
		name       string
		makeClient clientFactory
		wantErr    bool
	}{
		{
			name: "download images",
			makeClient: func(t *testing.T) client {
				ctrl := gomock.NewController(t)
				apiMock := NewMockAPI(ctrl)

				post1Images := []Image{
					{
						ID:          "image1",
						Extension:   "jpg",
						OriginalURL: "url1",
					},
				}

				apiMock.
					EXPECT().
					ListCreator(gomock.Any(), gomock.Any()).
					Times(1).
					Return(&ListCreator{
						Body: ListCreatorBody{
							Items: []Post{
								{
									ID:                "post1",
									Title:             "title1",
									PublishedDateTime: "2006-01-02T15:04:05+07:00",
									Body: &PostBody{
										Images: &post1Images,
									},
								},
							},
						},
					}, nil)

				for i := 0; i < len(post1Images); i++ {
					apiMock.
						EXPECT().
						Request(gomock.Any(), gomock.Eq(http.MethodGet), gomock.Any()).
						Times(1).
						Return(&http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(""))),
						}, nil)
				}

				storageMock := NewMockStorage(ctrl)

				storageMock.
					EXPECT().
					Exist(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(false, nil)

				storageMock.
					EXPECT().
					Save(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)

				return client{
					api:     apiMock,
					storage: storageMock,
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.makeClient(t)
			if err := c.Run(context.TODO(), "creator1"); (err != nil) != tt.wantErr {
				t.Errorf("client.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
