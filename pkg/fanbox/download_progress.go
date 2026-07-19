package fanbox

import (
	"fmt"
	"io"
	"strings"
	"time"
)

const downloadProgressInterval = time.Second
const downloadProgressBarWidth = 24

type downloadProgressBar struct {
	writer      io.Writer
	lastLineLen int
	displayed   bool
}

func newDownloadProgressBar(writer io.Writer) *downloadProgressBar {
	return &downloadProgressBar{writer: writer}
}

func (b *downloadProgressBar) Update(downloaded, total int64, elapsed time.Duration) {
	if b.writer == nil {
		return
	}

	line := formatDownloadProgress(downloaded, total, elapsed)
	padding := ""
	if len(line) < b.lastLineLen {
		padding = strings.Repeat(" ", b.lastLineLen-len(line))
	}
	_, _ = fmt.Fprintf(b.writer, "\r%s%s", line, padding)
	b.lastLineLen = len(line)
	b.displayed = true
}

func (b *downloadProgressBar) Finish() {
	if b.writer != nil && b.displayed {
		_, _ = fmt.Fprintln(b.writer)
	}
}

func (b *downloadProgressBar) Clear() {
	if b.writer != nil && b.displayed {
		_, _ = fmt.Fprintf(b.writer, "\r%s\r", strings.Repeat(" ", b.lastLineLen))
	}
}

func formatDownloadProgress(downloaded, total int64, elapsed time.Duration) string {
	if total <= 0 {
		return fmt.Sprintf("[ downloading ] %s  %s", formatByteCount(downloaded), formatByteRate(downloaded, elapsed))
	}

	percent := float64(downloaded) / float64(total)
	if percent > 1 {
		percent = 1
	}
	filled := int(percent * downloadProgressBarWidth)
	bar := strings.Repeat("#", filled) + strings.Repeat("-", downloadProgressBarWidth-filled)
	return fmt.Sprintf("[%s] %5.1f%%  %s / %s  %s",
		bar,
		percent*100,
		formatByteCount(downloaded),
		formatByteCount(total),
		formatByteRate(downloaded, elapsed),
	)
}

type downloadProgressReader struct {
	reader         io.Reader
	total          int64
	downloaded     int64
	startedAt      time.Time
	lastReportedAt time.Time
	now            func() time.Time
	report         func(downloaded, total int64, elapsed time.Duration)
	reportedDone   bool
}

func newDownloadProgressReader(reader io.Reader, total int64, report func(int64, int64, time.Duration)) *downloadProgressReader {
	now := time.Now()
	return &downloadProgressReader{
		reader:         reader,
		total:          total,
		startedAt:      now,
		lastReportedAt: now,
		now:            time.Now,
		report:         report,
	}
}

func (r *downloadProgressReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.downloaded += int64(n)

	now := r.now()
	done := err == io.EOF || (r.total > 0 && r.downloaded >= r.total)
	if (n > 0 || (done && r.downloaded > 0)) && (done || now.Sub(r.lastReportedAt) >= downloadProgressInterval) && !(done && r.reportedDone) {
		r.report(r.downloaded, r.total, now.Sub(r.startedAt))
		r.lastReportedAt = now
		r.reportedDone = done
	}

	return n, err
}

func formatByteCount(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	value := float64(bytes)
	units := []string{"KiB", "MiB", "GiB", "TiB"}
	for _, suffix := range units {
		value /= unit
		if value < unit || suffix == units[len(units)-1] {
			return fmt.Sprintf("%.1f %s", value, suffix)
		}
	}
	return ""
}

func formatByteRate(bytes int64, elapsed time.Duration) string {
	if elapsed <= 0 {
		return "0 B/s"
	}
	return formatByteCount(int64(float64(bytes)/elapsed.Seconds())) + "/s"
}
