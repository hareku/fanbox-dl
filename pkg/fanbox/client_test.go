package fanbox_test

import (
	"context"
	"fmt"
	"os"
	"sort"
	"testing"

	"github.com/hareku/fanbox-dl/pkg/fanbox"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/mod/sumdb/dirhash"
)

func TestClient_Run(t *testing.T) {
	type config struct {
		dryRun    bool
		skipFiles bool
		dirByPost bool
		dirByPlan bool
	}

	tests := []struct {
		config    config
		wantFiles []string
	}{
		{
			config: config{},
			wantFiles: []string{
				"oneshotatenno/2022-03-15-multiple-images-0-JF8xFtFv8uoQG2k7DS8Qg1rn.jpeg",
				"oneshotatenno/2022-03-15-multiple-images-1-UpT650o3tb4YSc4BCG28FMv5.jpeg",
				"oneshotatenno/2022-03-17-images-files-texts-0-RjuA09eUKC7dbs6F5V7gzn2y.jpeg",
				"oneshotatenno/2022-03-17-images-files-texts-1-M9bEuTJ9j3Rfp3xtpZf0RDeL.jpeg",
				"oneshotatenno/2022-03-17-images-files-texts-file-0-3mj6rAzFLrhm197FetXpMdFb.jpeg",
				"oneshotatenno/2022-03-17-images-files-texts-file-1-wc16n4QerQ8qxIBJrCbWWX3d.jpeg",
				"oneshotatenno/2022-03-17-multiple-files-file-0-SPyMpjKtXR20vrcHLu1jRu54.jpeg",
				"oneshotatenno/2022-03-17-multiple-files-file-1-9ZTsPyENS1e21anUrmaFW9Nl.jpeg",
			},
		},
		{
			config: config{
				dirByPost: true,
			},
			wantFiles: []string{
				"oneshotatenno/2022-03-15-multiple-images/0-JF8xFtFv8uoQG2k7DS8Qg1rn.jpeg",
				"oneshotatenno/2022-03-15-multiple-images/1-UpT650o3tb4YSc4BCG28FMv5.jpeg",
				"oneshotatenno/2022-03-17-images-files-texts/0-RjuA09eUKC7dbs6F5V7gzn2y.jpeg",
				"oneshotatenno/2022-03-17-images-files-texts/1-M9bEuTJ9j3Rfp3xtpZf0RDeL.jpeg",
				"oneshotatenno/2022-03-17-images-files-texts/file-0-3mj6rAzFLrhm197FetXpMdFb.jpeg",
				"oneshotatenno/2022-03-17-images-files-texts/file-1-wc16n4QerQ8qxIBJrCbWWX3d.jpeg",
				"oneshotatenno/2022-03-17-multiple-files/file-0-SPyMpjKtXR20vrcHLu1jRu54.jpeg",
				"oneshotatenno/2022-03-17-multiple-files/file-1-9ZTsPyENS1e21anUrmaFW9Nl.jpeg",
			},
		},
		{
			config: config{
				dirByPlan: true,
			},
			wantFiles: []string{
				"oneshotatenno/0yen/2022-03-15-multiple-images-0-JF8xFtFv8uoQG2k7DS8Qg1rn.jpeg",
				"oneshotatenno/0yen/2022-03-15-multiple-images-1-UpT650o3tb4YSc4BCG28FMv5.jpeg",
				"oneshotatenno/0yen/2022-03-17-images-files-texts-0-RjuA09eUKC7dbs6F5V7gzn2y.jpeg",
				"oneshotatenno/0yen/2022-03-17-images-files-texts-1-M9bEuTJ9j3Rfp3xtpZf0RDeL.jpeg",
				"oneshotatenno/0yen/2022-03-17-images-files-texts-file-0-3mj6rAzFLrhm197FetXpMdFb.jpeg",
				"oneshotatenno/0yen/2022-03-17-images-files-texts-file-1-wc16n4QerQ8qxIBJrCbWWX3d.jpeg",
				"oneshotatenno/0yen/2022-03-17-multiple-files-file-0-SPyMpjKtXR20vrcHLu1jRu54.jpeg",
				"oneshotatenno/0yen/2022-03-17-multiple-files-file-1-9ZTsPyENS1e21anUrmaFW9Nl.jpeg",
			},
		},
		{
			config: config{
				dirByPost: true,
				dirByPlan: true,
			},
			wantFiles: []string{
				"oneshotatenno/0yen/2022-03-15-multiple-images/0-JF8xFtFv8uoQG2k7DS8Qg1rn.jpeg",
				"oneshotatenno/0yen/2022-03-15-multiple-images/1-UpT650o3tb4YSc4BCG28FMv5.jpeg",
				"oneshotatenno/0yen/2022-03-17-images-files-texts/0-RjuA09eUKC7dbs6F5V7gzn2y.jpeg",
				"oneshotatenno/0yen/2022-03-17-images-files-texts/1-M9bEuTJ9j3Rfp3xtpZf0RDeL.jpeg",
				"oneshotatenno/0yen/2022-03-17-images-files-texts/file-0-3mj6rAzFLrhm197FetXpMdFb.jpeg",
				"oneshotatenno/0yen/2022-03-17-images-files-texts/file-1-wc16n4QerQ8qxIBJrCbWWX3d.jpeg",
				"oneshotatenno/0yen/2022-03-17-multiple-files/file-0-SPyMpjKtXR20vrcHLu1jRu54.jpeg",
				"oneshotatenno/0yen/2022-03-17-multiple-files/file-1-9ZTsPyENS1e21anUrmaFW9Nl.jpeg",
			},
		},
		{
			config: config{
				skipFiles: true,
				dirByPost: true,
			},
			wantFiles: []string{
				"oneshotatenno/2022-03-15-multiple-images/0-JF8xFtFv8uoQG2k7DS8Qg1rn.jpeg",
				"oneshotatenno/2022-03-15-multiple-images/1-UpT650o3tb4YSc4BCG28FMv5.jpeg",
				"oneshotatenno/2022-03-17-images-files-texts/0-RjuA09eUKC7dbs6F5V7gzn2y.jpeg",
				"oneshotatenno/2022-03-17-images-files-texts/1-M9bEuTJ9j3Rfp3xtpZf0RDeL.jpeg",
				// "oneshotatenno/2022-03-17-images-files-texts/file-0-3mj6rAzFLrhm197FetXpMdFb.jpeg",
				// "oneshotatenno/2022-03-17-images-files-texts/file-1-wc16n4QerQ8qxIBJrCbWWX3d.jpeg",
				// "oneshotatenno/2022-03-17-multiple-files/file-0-SPyMpjKtXR20vrcHLu1jRu54.jpeg",
				// "oneshotatenno/2022-03-17-multiple-files/file-1-9ZTsPyENS1e21anUrmaFW9Nl.jpeg",
			},
		},
		{
			config: config{
				dryRun:    true,
				skipFiles: true,
				dirByPost: true,
			},
			wantFiles: nil,
		},
	}

	logger := fanbox.NewLogger(&fanbox.NewLoggerInput{
		Out:     os.Stdout,
		Verbose: true,
	})
	httpClient := retryablehttp.NewClient()
	httpClient.Logger = logger

	for _, tt := range tests {
		t.Run(fmt.Sprintf("config:%+v", tt.config), func(t *testing.T) {
			saveDir, err := os.MkdirTemp("", "fanbox-dl-testing-")
			require.NoError(t, err)
			t.Cleanup(func() {
				assert.NoError(t, os.RemoveAll(saveDir))
			})

			client := fanbox.Client{
				CheckAllPosts: true,
				DryRun:        tt.config.dryRun,
				SkipFiles:     tt.config.skipFiles,
				OfficialAPIClient: &fanbox.OfficialAPIClient{
					HTTPClient: httpClient,
				},
				Storage: &fanbox.LocalStorage{
					SaveDir:   saveDir,
					DirByPost: tt.config.dirByPost,
					DirByPlan: tt.config.dirByPlan,
				},
				Logger: logger,
			}

			require.NoError(t, client.Run(context.Background(), "oneshotatenno"))

			got, err := dirhash.DirFiles(saveDir, "")
			require.NoError(t, err)

			sort.Strings(tt.wantFiles)
			sort.Strings(got)
			assert.Equal(t, tt.wantFiles, got)
		})
	}
}
