package fman

import (
	"io"

	"github.com/nvthongswansea/xtreme/internal/models"
)

// FmanUsecase provides an interface for interacting with file.
type FmanUsecase interface {
	// Update a file.
	UploadFile(newFile models.File, contentReader io.Reader) error

	// Copy a file.
	CopyFile()

	// Move a file.
	MoveFile()

	// Remove a file.
	RemoveFile()

	// Move a file to recycle bin.
	MoveFileToRecyleBin()
}