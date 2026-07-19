package fanbox

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"
)

func TestDownloadProgressReaderReportsProgressAndCompletion(t *testing.T) {
	var reports []int64
	reader := newDownloadProgressReader(strings.NewReader("abcdef"), 6, func(downloaded, total int64, elapsed time.Duration) {
		if total != 6 {
			t.Errorf("total = %d, want 6", total)
		}
		reports = append(reports, downloaded)
	})

	now := reader.startedAt
	reader.now = func() time.Time {
		now = now.Add(time.Second)
		return now
	}

	got, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read progress reader: %v", err)
	}
	if string(got) != "abcdef" {
		t.Errorf("content = %q, want %q", got, "abcdef")
	}
	if len(reports) == 0 || reports[len(reports)-1] != 6 {
		t.Fatalf("reports = %v, want final report at 6 bytes", reports)
	}
}

func TestFormatByteCount(t *testing.T) {
	tests := map[int64]string{
		512:       "512 B",
		1024:      "1.0 KiB",
		128 << 20: "128.0 MiB",
		3 << 30:   "3.0 GiB",
	}
	for input, want := range tests {
		if got := formatByteCount(input); got != want {
			t.Errorf("formatByteCount(%d) = %q, want %q", input, got, want)
		}
	}
}

func TestDownloadProgressReaderReportsCompletionWithoutContentLength(t *testing.T) {
	var final int64
	reader := newDownloadProgressReader(strings.NewReader("abc"), -1, func(downloaded, total int64, elapsed time.Duration) {
		final = downloaded
	})
	reader.now = func() time.Time { return reader.startedAt }

	if _, err := io.ReadAll(reader); err != nil {
		t.Fatalf("read progress reader: %v", err)
	}
	if final != 3 {
		t.Errorf("final report = %d, want 3", final)
	}
}

func TestDownloadProgressBarRewritesOneTerminalLine(t *testing.T) {
	var output bytes.Buffer
	bar := newDownloadProgressBar(&output)
	bar.Update(25, 100, time.Second)
	bar.Update(100, 100, 2*time.Second)
	bar.Finish()

	got := output.String()
	if strings.Count(got, "\n") != 1 {
		t.Fatalf("newline count = %d, want 1; output = %q", strings.Count(got, "\n"), got)
	}
	if strings.Count(got, "\r") != 2 {
		t.Fatalf("carriage return count = %d, want 2; output = %q", strings.Count(got, "\r"), got)
	}
	if !strings.Contains(got, "100.0%") || !strings.Contains(got, "50 B/s") {
		t.Errorf("final progress output = %q", got)
	}
}

func TestDownloadProgressBarClear(t *testing.T) {
	var output bytes.Buffer
	bar := newDownloadProgressBar(&output)
	bar.Update(1, 10, time.Second)
	bar.Clear()

	if !strings.HasSuffix(output.String(), "\r") {
		t.Errorf("cleared output = %q, want trailing carriage return", output.String())
	}
	if strings.Contains(output.String(), "\n") {
		t.Errorf("cleared output = %q, want no newline", output.String())
	}
}
