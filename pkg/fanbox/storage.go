package fanbox

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/hareku/go-filename"
	"github.com/hareku/go-strlimit"
)

//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

type Storage interface {
	// Save saves the file.
	Save(post Post, order int, img Image, file io.Reader) error

	// Exist returns whether the image is already saved.
	Exist(post Post, order int, img Image) (bool, error)
}

type localStorage struct {
	saveDir   string
	dirByPost bool
}

type NewLocalStorageInput struct {
	SaveDir   string
	DirByPost bool
}

func NewLocalStorage(i *NewLocalStorageInput) Storage {
	return &localStorage{
		saveDir:   i.SaveDir,
		dirByPost: i.DirByPost,
	}
}

func (s *localStorage) Save(post Post, order int, img Image, reader io.Reader) error {
	name := s.makeFileName(post, order, img)

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

	_, err = io.Copy(file, reader)
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

func (s *localStorage) Exist(post Post, order int, img Image) (bool, error) {
	_, err := os.Stat(s.makeFileName(post, order, img))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to stat file: %w", err)
	}

	return true, nil
}

// limitOsSafely limits the string length for OS safely.
func (s *localStorage) limitOsSafely(name string) string {
	switch runtime.GOOS {
	case "windows":
		return strlimit.LimitRunesWithEnd(name, 210, "...")
	default:
		return strlimit.LimitBytesWithEnd(name, 250, "...")
	}
}

func (s *localStorage) makeFileName(post Post, order int, img Image) string {
	date, err := time.Parse(time.RFC3339, post.PublishedDateTime)
	if err != nil {
		panic(fmt.Errorf("failed to parse post published date time %s: %w", post.PublishedDateTime, err))
	}

	title := filename.EscapeString(post.Title, "-")

	if s.dirByPost {
		// [SaveDirectory]/[UserID]/2006-01-02-[Post Title]/[Order]-[Image ID].[Image Extension]
		return filepath.Join(s.saveDir, post.CreatorID, s.limitOsSafely(fmt.Sprintf("%s-%s", date.UTC().Format("2006-01-02"), title)), fmt.Sprintf("%d-%s.%s", order, img.ID, img.Extension))
	}

	// [SaveDirectory]/[UserID]/2006-01-02-[Post Title]-[Order]-[Image ID].[Image Extension]
	return filepath.Join(s.saveDir, post.CreatorID, fmt.Sprintf("%s.%s", s.limitOsSafely(fmt.Sprintf("%s-%s-%d-%s", date.UTC().Format("2006-01-02"), title, order, img.ID)), img.Extension))
}
