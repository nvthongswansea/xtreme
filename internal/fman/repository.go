package fman

import "github.com/nvthongswansea/xtreme/internal/models"

// FManDBRepo provides an interface for operations on file manager database.
type FManDBRepo interface {
	// InsertFileRecord inserts a file record to db. Returns the newly inserted row, if success;
	// otherwise return an error.
	InsertFileRecord(newFile models.File) (models.File, error)

	// ReadFileRecord reads a file record from the db with a given UUID.
	ReadFileRecord(UUID string) (models.File, error)

	// UpdateFileRecord updates a file record in the db.
	UpdateFileRecord(file models.File) error

	// SoftRemoveFileRecord flags a file record as deleted file.
	// e.g. set `is_deleted` field to true.
	SoftRemoveFileRecord(UUID string) error

	// HardRemoveFileRecord removes a file record completely from the db.
	HardRemoveFileRecord(UUID string) error
}
