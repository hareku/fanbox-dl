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
	mock "github.com/hareku/fanbox-dl/pkg/fanbox/mock"
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
				apiClientMock := mock.NewMockApiClient(ctrl)

				post1Images := []Image{
					{
						ID:          "image1",
						Extension:   "jpg",
						OriginalURL: "url1",
					},
				}

				apiClientMock.
					EXPECT().
					RequestAsJSON(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(1).
					DoAndReturn(func(_ context.Context, url string, list *ListCreator) error {
						*list = ListCreator{
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
						}

						return nil
					})

				for i := 0; i < len(post1Images); i++ {
					apiClientMock.
						EXPECT().
						Request(gomock.Any(), gomock.Any()).
						Times(1).
						Return(&http.Response{
							StatusCode: 200,
							Body:       io.NopCloser(bytes.NewReader([]byte(""))),
						}, nil)
				}

				fileClientMock := mock.NewMockFileClient(ctrl)

				fileClientMock.
					EXPECT().
					DoesExist(gomock.Any()).
					Times(1).
					Return(false, nil)

				fileClientMock.
					EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Times(1).
					Return(nil)

				return client{
					userID:     "user1",
					saveDir:    TempDir,
					apiClient:  apiClientMock,
					fileClient: fileClientMock,
				}
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.makeClient(t)
			if err := c.Run(context.TODO()); (err != nil) != tt.wantErr {
				t.Errorf("client.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
