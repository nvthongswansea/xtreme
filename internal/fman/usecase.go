package fman

import (
	"io"
)

// FmanUsecase provides an interface for interacting with file.
type FmanUsecase interface {
	// Update a file.
	UploadFile(filename, parentUUID string, contentReader io.Reader) error

	// Copy a file to a new location.
	CopyFile(srcUUID, dstParentUUID string) error

	// Create a new directory/folder.
	CreateNewDirectory(dirname, parentUUID string) error

	// Move a file.
	MoveFile()

	// Remove a file.
	RemoveFile()

	// Move a file to recycle bin.
	MoveFileToRecyleBin()
}
