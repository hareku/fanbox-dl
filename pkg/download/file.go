package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func (c *Client) makeFileName(post Post, order int, img Image) string {
	date, err := time.Parse(time.RFC3339, post.PublishedDateTime)
	if err != nil {
		panic(fmt.Errorf("failed to parse post published date time %s: %w", post.PublishedDateTime, err))
	}

	if c.SeparateByPost {
		// [SaveDirectory]/2006-01-02-[Post Title]/[Order]-[Image ID].[Image Extension]
		return filepath.Join(c.SaveDir, fmt.Sprintf("%s-%s", date.UTC().Format("2006-01-02"), post.Title), fmt.Sprintf("%d-%s.%s", order, img.ID, img.Extension))
	}

	// [SaveDirectory]/2006-01-02-[Post Title]-[Order]-[Image ID].[Image Extension]
	return filepath.Join(c.SaveDir, fmt.Sprintf("%s-%s-%d-%s.%s", date.UTC().Format("2006-01-02"), post.Title, order, img.ID, img.Extension))
}

func (c *Client) saveFile(name string, resp *http.Response) error {
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
		// Remove crashed file
		errRemove := os.Remove(file.Name())
		if errRemove != nil {
			return fmt.Errorf("file copying error and couldn't remove a crashed file (%s): %w", file.Name(), errRemove)
		}

		return fmt.Errorf("file copying error: %w", err)
	}

	return nil
}

func (c *Client) isDownloaded(name string) bool {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false
	}

	return true
}
