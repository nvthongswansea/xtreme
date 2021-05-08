package fileUtils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// FileSaveReadRemover provides an interface to save/read/remove a file to/from/from a source.
type FileSaveReadRemover interface {
	// Save file to a source.
	SaveFile(relFilePath string, contentReadCloser io.ReadCloser) (int64, string, error)

	// ReadFile returns an instance of io.ReadCloser. Data can be read from the instance via
	// Read() function. NOTE: Remember to Close() after reading the content.
	ReadFile(relFilePath string) (io.ReadCloser, error)

	// Remove a file from a source.
	RemoveFile(relFilePath string) error
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

// SaveFile saves a file from a reader to the local disk, return the number of bytes
// saved on the local disk and the location of the file.
// If the filename already exists, return error.
func (fs *LocalFileOperator) SaveFile(relFilePath string, contentReadCloser io.ReadCloser) (int64, string, error) {
	defer contentReadCloser.Close()
	// absolute filepath on disk.
	absFilePathOD := filepath.Join(fs.basePath, relFilePath)
	// Check if the file already exists.
	if _, err := os.Stat(absFilePathOD); err == nil {
		// Return error if the file already exists.
		return 0, "", fmt.Errorf("File %s already exist", absFilePathOD)
	}
	// Create a new empty dst file.
	dstF, err := os.Create(absFilePathOD)
	if err != nil {
		return 0, "", err
	}
	defer dstF.Close()
	// Copy content to the dst file.
	size, err := io.Copy(dstF, contentReadCloser)
	return size, absFilePathOD, err
}

// ReadFile returns an os.File pointer with a given filename, which can be only used for reading the file content from the
// local storage.
func (fs *LocalFileOperator) ReadFile(relFilePath string) (io.ReadCloser, error) {
	// absolute filepath on disk.
	absFilePathOD := filepath.Join(fs.basePath, relFilePath)
	return os.Open(absFilePathOD)
}

// RemoveFile removes a file from the local disk.
func (fs *LocalFileOperator) RemoveFile(relFilePath string) error {
	// absolute filepath on disk.
	absFilePathOD := filepath.Join(fs.basePath, relFilePath)
	return os.Remove(absFilePathOD)
}
