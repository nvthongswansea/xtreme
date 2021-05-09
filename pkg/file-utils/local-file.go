package fileUtils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const filenameInvalidChars string = "\\/.?%*:|\"<>,;= "

// FileSaveReadRemover provides an interface to save/read/remove a file to/from/from a source.
type FileSaveReadRemover interface {
	// SaveCloseFile Save file to a source, then close contentReadCloser.
	SaveCloseFile(relFilePath string, contentReadCloser io.ReadCloser) (int64, string, error)

	// ReadFile returns an instance of io.ReadCloser. Data can be read from the instance via
	// Read() function. NOTE: Remember to Close() after reading the content.
	ReadFile(absFilePathOD string) (io.ReadCloser, error)

	// RemoveFile removes a file from a source.
	RemoveFile(absFilePathOD string) error
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

// SaveCloseFile saves a file from a reader to the local disk, and close
// contentReadCloser. Return the number of bytes
// saved on the local disk and the absolute location of the file.
// If the filename already exists, return error.
func (fs *LocalFileOperator) SaveCloseFile(relFilePath string, contentReadCloser io.ReadCloser) (int64, string, error) {
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
func (fs *LocalFileOperator) ReadFile(absFilePathOD string) (io.ReadCloser, error) {
	return os.Open(absFilePathOD)
}

// RemoveFile removes a file from the local disk.
func (fs *LocalFileOperator) RemoveFile(absFilePathOD string) error {
	return os.Remove(absFilePathOD)
}

// FilenameValidator is signature of a func to validate the filename.
type FilenameValidator func(filename string) bool

// IsFilenameOk checks if filename is valid. If the filename doesn't
// contain any prohibited characters, return true; otherwise
// return false
func IsFilenameOk(filename string) bool {
	if filename == "" || strings.ContainsAny(filename, filenameInvalidChars) {
		return false
	}
	return true
}
