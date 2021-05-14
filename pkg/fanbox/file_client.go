package fanbox

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type FileClient interface {
	// Save saves the passed reader as a file.
	Save(name string, reader io.Reader) error

	// DoesExist returns whether the given name file exist.
	DoesExist(name string) (bool, error)
}

type flieClient struct{}

// NewFileClient returns the new FileClient instance.
func NewFileClient() FileClient {
	return &flieClient{}
}

func (c *flieClient) Save(name string, reader io.Reader) error {
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

func (c *flieClient) DoesExist(name string) (bool, error) {
	_, err := os.Stat(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to stat file: %w", err)
	}

	return true, nil
}
