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
	SaveFile(filename string, contentReadCloser io.ReadCloser) (int64, string, error)

	// ReadFile returns an instance of io.ReadCloser. Data can be read from the instance via
	// Read() function. NOTE: Remember to Close() after reading the content.
	ReadFile(filename string) (io.ReadCloser, error)

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

// SaveFile saves a file from a reader to the local disk, return the number of bytes
// saved on the local disk and the location of the file.
// If the filename already exists, return error.
func (fs *LocalFileOperator) SaveFile(filename string, contentReadCloser io.ReadCloser) (int64, string, error) {
	defer contentReadCloser.Close()
	// filePathOD filepath on disk.
	filePathOD := filepath.Join(fs.basePath, filename)
	// Check if the file already exists.
	if _, err := os.Stat(filePathOD); err == nil {
		// Return error if the file already exists.
		return 0, "", fmt.Errorf("File %s already exist", filePathOD)
	}
	// Create a new empty dst file.
	dstF, err := os.Create(filePathOD)
	if err != nil {
		return 0, "", err
	}
	defer dstF.Close()
	// Copy content to the dst file.
	size, err := io.Copy(dstF, contentReadCloser)
	return size, filePathOD, err
}

// ReadFile returns an os.File pointer with a given filename, which can be only used for reading the file content from the
// local storage.
func (fs *LocalFileOperator) ReadFile(filename string) (io.ReadCloser, error) {
	// filePathOD filepath on disk.
	filePathOD := filepath.Join(fs.basePath, filename)
	return os.Open(filePathOD)
}

// RemoveFile removes a file from the local disk.
func (fs *LocalFileOperator) RemoveFile(filename string) error {
	// filePathOD filepath on disk.
	filePathOD := filepath.Join(fs.basePath, filename)
	return os.Remove(filePathOD)
}
