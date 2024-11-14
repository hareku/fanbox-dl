package fanbox

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode"

	"github.com/hareku/go-filename"
	"github.com/hareku/go-strlimit"
)

type LocalStorage struct {
	SaveDir   string
	DirByPost bool
	DirByPlan bool

	RemoveUnprintableChars bool
}

func (s *LocalStorage) Save(post Post, order int, d Downloadable, r io.Reader) error {
	name := s.makeFileName(post, order, d)

	dir := filepath.Dir(name)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0775)
		if err != nil {
			return fmt.Errorf("create a directory (%s): %w", dir, err)
		}
	}

	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0775)
	if err != nil {
		return fmt.Errorf("open a file: %w", err)
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

func (s *LocalStorage) Exist(post Post, order int, d Downloadable) (bool, error) {
	_, err := os.Stat(s.makeFileName(post, order, d))
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("stat file: %w", err)
	}

	return true, nil
}

// limitOsSafely limits the string length for OS safely.
func (s *LocalStorage) limitOsSafely(name string) string {
	switch runtime.GOOS {
	case "windows":
		return strlimit.LimitRunesWithEnd(name, 210, "...")
	default:
		return strlimit.LimitBytesWithEnd(name, 250, "...")
	}
}

func (s *LocalStorage) makeFileName(post Post, order int, d Downloadable) string {
	date, err := time.Parse(time.RFC3339, post.PublishedDateTime)
	if err != nil {
		panic(fmt.Errorf("parse post published date time %s: %w", post.PublishedDateTime, err))
	}

	title := strings.TrimSpace(filename.EscapeString(post.Title, "-"))
	if s.RemoveUnprintableChars {
		title = strings.Map(func(r rune) rune {
			if unicode.IsPrint(r) {
				return r
			}
			return -1
		}, title)
	}

	fileType := ""
	// for backward-compatibility, insert "-file-" identifier
	if _, ok := d.(File); ok {
		fileType = "file-"
	}

	planDir := ""
	if s.DirByPlan {
		planDir = fmt.Sprintf("%dyen", post.FeeRequired)
	}

	if s.DirByPost {
		// [SaveDirectory]/[CreatorID]/2006-01-02-[Post Title]/[Order]-[ID].[Extension]
		return filepath.Join(
			s.SaveDir,
			post.CreatorID,
			planDir,
			s.limitOsSafely(fmt.Sprintf("%s-%s", date.UTC().Format("2006-01-02"), title)),
			fmt.Sprintf("%s%d-%s.%s", fileType, order, d.GetID(), d.GetExtension()),
		)
	}

	// [SaveDirectory]/[CreatorID]/2006-01-02-[Post Title]-[Order]-[ID].[Extension]
	return filepath.Join(
		s.SaveDir,
		post.CreatorID,
		planDir,
		fmt.Sprintf(
			"%s.%s",
			s.limitOsSafely(
				fmt.Sprintf(
					"%s-%s-%s%d-%s",
					date.UTC().Format("2006-01-02"),
					title,
					fileType,
					order,
					d.GetID(),
				),
			),
			d.GetExtension(),
		),
	)
}
