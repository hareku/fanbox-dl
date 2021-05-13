package fanbox

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/hareku/go-filename"
	"github.com/hareku/go-strlimit"
)

func osSafeLimit(s string) string {
	switch runtime.GOOS {
	case "windows":
		return strlimit.LimitRunesWithEnd(s, 210, "...")
	default:
		return strlimit.LimitBytesWithEnd(s, 250, "...")
	}
}

func (c *client) makeFileName(post Post, order int, img Image) string {
	date, err := time.Parse(time.RFC3339, post.PublishedDateTime)
	if err != nil {
		panic(fmt.Errorf("failed to parse post published date time %s: %w", post.PublishedDateTime, err))
	}

	title := filename.EscapeString(post.Title, "-")

	if c.separateByPost {
		// [SaveDirectory]/[UserID]/2006-01-02-[Post Title]/[Order]-[Image ID].[Image Extension]
		return filepath.Join(c.saveDir, c.userID, osSafeLimit(fmt.Sprintf("%s-%s", date.UTC().Format("2006-01-02"), title)), fmt.Sprintf("%d-%s.%s", order, img.ID, img.Extension))
	}

	// [SaveDirectory]/[UserID]/2006-01-02-[Post Title]-[Order]-[Image ID].[Image Extension]
	return filepath.Join(c.saveDir, c.userID, fmt.Sprintf("%s.%s", osSafeLimit(fmt.Sprintf("%s-%s-%d-%s", date.UTC().Format("2006-01-02"), title, order, img.ID)), img.Extension))
}

func (c *client) saveFile(name string, resp *http.Response) error {
	dir := filepath.Dir(name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0775)
		if err != nil {
			return fmt.Errorf("failed to create a directory (%s): %w", dir, err)
		}
	}

	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0775)
	if err != nil {
		return fmt.Errorf("failed to open a file: %w", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		// Remove the crashed file
		fileName := file.Name()
		file.Close()

		if removeRrr := os.Remove(fileName); removeRrr != nil {
			return fmt.Errorf("file copying error and couldn't remove a crashed file (%s): %w", file.Name(), removeRrr)
		}

		return fmt.Errorf("file copying error: %w", err)
	}

	return nil
}

func (c *client) isDownloaded(name string) (bool, error) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to stat file: %w", err)
	}

	return true, nil
}
