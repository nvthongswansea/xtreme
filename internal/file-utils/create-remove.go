package fileUtils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileSaverRemover provides an interface to save/remove a file to/from the source.
type FileSaverRemover interface {
	// Save file to a source.
	SaveFile(filename string, contentReader io.Reader) error
	// Remove a file from a source.
	RemoveFile(filename string) error
}

type LocalFileOperator struct {
	basePath string
}

// CreateNewLocalFileOperator create a new LocalFileOperator
func CreateNewLocalFileOperator(basePath string) *LocalFileOperator {
	return &LocalFileOperator{
		basePath,
	}
}

// SaveFile saves a file from a reader to the local disk. If the filename already exists,
// return error.
func (fs *LocalFileOperator) SaveFile(filename string, contentReader io.Reader) error {
	// filePathOD filepath on disk.
	filePathOD := filepath.Join(fs.basePath, filename)
	// Check if the file already exists.
	if _, err := os.Stat(filePathOD); !os.IsNotExist(err) {
		// Return error if the file already exists.
		return fmt.Errorf("File %s already exist", filePathOD)
	}
	// Create a new empty dst file.
	dstF, err := os.Create(filePathOD)
	if err != nil {
		return err
	}
	defer dstF.Close()
	// Copy content to the dst file.
	_, err = io.Copy(dstF, contentReader)
	return err
}

// RemoveFile removes a fiel from the local disk.
func (fs *LocalFileOperator) RemoveFile(filename string) error {
	return nil
}
