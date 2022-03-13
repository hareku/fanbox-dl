package fanbox

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hareku/go-filename"
	"github.com/hareku/go-strlimit"
)

//go:generate mockgen -source=$GOFILE -destination=mock_$GOFILE -package=$GOPACKAGE

type FileStorage interface {
	// Save saves the file.
	Save(post PostInfoBody, order int, file File, r io.Reader) error

	// Exist returns whether the file is already saved.
	Exist(post PostInfoBody, order int, file File) (bool, error)
}

type localFileStorage struct {
	saveDir   string
	dirByPost bool
}

type NewLocalFileStorageInput struct {
	SaveDir   string
	DirByPost bool
}

func NewLocalFileStorage(i *NewLocalFileStorageInput) FileStorage {
	return &localFileStorage{
		saveDir:   i.SaveDir,
		dirByPost: i.DirByPost,
	}
}

func (s *localFileStorage) Save(post PostInfoBody, order int, f File, r io.Reader) error {
	name := s.makeFileName(post, order, f)

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

	_, err = io.Copy(file, r)
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

func (s *localFileStorage) Exist(post PostInfoBody, order int, f File) (bool, error) {
	_, err := os.Stat(s.makeFileName(post, order, f))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to stat file: %w", err)
	}

	return true, nil
}

// limitOsSafely limits the string length for OS safely.
func (s *localFileStorage) limitOsSafely(name string) string {
	switch runtime.GOOS {
	case "windows":
		return strlimit.LimitRunesWithEnd(name, 210, "...")
	default:
		return strlimit.LimitBytesWithEnd(name, 250, "...")
	}
}

func (s *localFileStorage) makeFileName(post PostInfoBody, order int, f File) string {
	date, err := time.Parse(time.RFC3339, post.PublishedDateTime)
	if err != nil {
		panic(fmt.Errorf("failed to parse post published date time %s: %w", post.PublishedDateTime, err))
	}

	title := strings.TrimSpace(filename.EscapeString(post.Title, "-"))

	if s.dirByPost {
		// [SaveDirectory]/[CreatorID]/2006-01-02-[Post Title]/[Order]-[Image ID].[Image Extension]
		return filepath.Join(
			s.saveDir,
			post.CreatorID,
			s.limitOsSafely(
				fmt.Sprintf("%s-%s", date.UTC().Format("2006-01-02"), title),
			),
			fmt.Sprintf("%d-%s.%s", order, f.ID, f.Extension))
	}

	// [SaveDirectory]/[CreatorID]/2006-01-02-[Post Title]-file-[Order]-[Image ID].[Image Extension]
	return filepath.Join(
		s.saveDir,
		post.CreatorID,
		fmt.Sprintf(
			"%s.%s",
			s.limitOsSafely(
				fmt.Sprintf("%s-%s-file-%d-%s",
					date.UTC().Format("2006-01-02"),
					title,
					order,
					f.ID)),
			f.Extension))
}
