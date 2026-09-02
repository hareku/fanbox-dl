package fanbox

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestClientShouldStopPagination(t *testing.T) {
	startDate := time.Date(2024, time.January, 10, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, time.January, 20, 23, 59, 59, 0, time.UTC)

	tests := []struct {
		name     string
		client   Client
		posts    []Post
		wantStop bool
	}{
		{
			name:     "oldest post is before start date",
			client:   Client{StartDate: &startDate, EndDate: &endDate},
			posts:    []Post{{PublishedDateTime: "2024-01-09T23:59:59Z"}},
			wantStop: true,
		},
		{
			name:     "oldest post equals start date",
			client:   Client{StartDate: &startDate, EndDate: &endDate},
			posts:    []Post{{PublishedDateTime: "2024-01-10T00:00:00Z"}},
			wantStop: false,
		},
		{
			name:     "end date alone does not stop pagination",
			client:   Client{EndDate: &endDate},
			posts:    []Post{{PublishedDateTime: "2024-01-01T00:00:00Z"}},
			wantStop: false,
		},
		{
			name:     "invalid post date does not stop pagination",
			client:   Client{StartDate: &startDate},
			posts:    []Post{{PublishedDateTime: "invalid"}},
			wantStop: false,
		},
		{
			name:     "empty page does not stop pagination",
			client:   Client{StartDate: &startDate},
			wantStop: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			content := ListCreatorResponse{Body: tt.posts}
			assert.Equal(t, tt.wantStop, tt.client.shouldStopPagination(&content))
		})
	}
}
