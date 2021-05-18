package file

import (
	"context"
	"github.com/nvthongswansea/xtreme/internal/models"
)

// Repository holds all operations on files.
type Repository interface {
	Inserter
	Reader
	Updater
	Remover
}

// Inserter holds inserting operations on files.
type Inserter interface {
	// InsertFile inserts a new file record to db.
	InsertFile(ctx context.Context, newFile models.File) error
}

// Reader holds reading operations on files.
type Reader interface {
	// GetFile gets a file record from the db with a given UUID.
	GetFile(ctx context.Context, fileUUID string) (models.File, error)

	// GetFileMetadata gets file metadata from the db with a given UUID.
	GetFileMetadata(ctx context.Context, fileUUID string) (models.FileMetadata, error)

	// GetFileMetadataBatch gets multiple files' metadata from the db with given UUIDs.
	GetFileMetadataBatch(ctx context.Context, fileUUIDs []string) ([]models.FileMetadata, error)

	// GetFileMetadataListByDir gets all child-files' metadata in a given directory.
	GetFileMetadataListByDir(ctx context.Context, dirUUID string) ([]models.FileMetadata, error)

	// GetFileMetadataListByName gets all files' metadata named after the given filename.
	GetFileMetadataListByName(ctx context.Context, filename, parentDirUUID string) ([]models.FileMetadata, error)

	// IsFilenameExist checks if a specific filename exists in a specific directory.
	IsFilenameExist(ctx context.Context, name, parentDirUUID string) (bool, error)
}

// Updater holds updating operations on files.
type Updater interface {
	// UpdateFilename updates name of a file in the db.
	UpdateFilename(ctx context.Context, fileUUID, newFilename string) error

	// UpdateParentDirUUID updates the parent dirUUID of a file.
	UpdateParentDirUUID(ctx context.Context, fileUUID, newParentDirUUID string) error
}

// Remover holds removing operations on files.
type Remover interface {
	// SoftRemoveFile flags a file record as deleted file.
	// e.g. set `is_deleted` field to true.
	SoftRemoveFile(ctx context.Context, fileUUID string) error

	// HardRemoveFile removes a file record completely from the db.
	HardRemoveFile(ctx context.Context, fileUUID string, fileRmCallback func(string) error) error
}