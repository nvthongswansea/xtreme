package file

import (
	"context"
	"github.com/nvthongswansea/xtreme/internal/models"
	"github.com/nvthongswansea/xtreme/internal/repository/transaction"
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
	InsertFile(ctx context.Context, tx transaction.RollbackCommitter, newFile models.File) (string, error)
}

// Reader holds reading operations on files.
type Reader interface {
	// GetFile gets a file record from the db with a given UserUUID.
	GetFile(ctx context.Context, tx transaction.RollbackCommitter, fileUUID string) (models.File, error)

	// GetFileMetadata gets file metadata from the db with a given UserUUID.
	GetFileMetadata(ctx context.Context, tx transaction.RollbackCommitter, fileUUID string) (models.FileMetadata, error)

	// GetFileMetadataBatch gets multiple files' metadata from the db with given UUIDs.
	GetFileMetadataBatch(ctx context.Context, tx transaction.RollbackCommitter, fileUUIDs []string) ([]models.FileMetadata, error)

	// GetFileMetadataListByDir gets all child-files' metadata in a given directory.
	GetFileMetadataListByDir(ctx context.Context, tx transaction.RollbackCommitter, dirUUID string) ([]models.FileMetadata, error)

	// GetFileMetadataListByName gets all files' metadata named after the given filename.
	GetFileMetadataListByName(ctx context.Context, tx transaction.RollbackCommitter, parentDirUUID, filename string) ([]models.FileMetadata, error)

	// IsFilenameExist checks if a specific filename exists in a specific directory.
	IsFilenameExist(ctx context.Context, tx transaction.RollbackCommitter, parentDirUUID, name string) (bool, error)
}

// Updater holds updating operations on files.
type Updater interface {
	// UpdateFilename updates name of a file in the db.
	UpdateFilename(ctx context.Context, tx transaction.RollbackCommitter, newFilename, fileUUID string) error

	// UpdateFileRelPathOD updates relative path of the file on the local storage.
	UpdateFileRelPathOD(ctx context.Context, tx transaction.RollbackCommitter, relPathOD, fileUUID string) error

	// UpdateFileSize updates size of a file.
	UpdateFileSize(ctx context.Context, tx transaction.RollbackCommitter, size int64, fileUUID string) error

	// UpdateParentDirUUID updates the parent dirUUID of a file.
	UpdateParentDirUUID(ctx context.Context, tx transaction.RollbackCommitter, newParentDirUUID, fileUUID string) error
}

// Remover holds removing operations on files.
type Remover interface {
	// SoftRemoveFile flags a file record as deleted file.
	// e.g. set `is_deleted` field to true.
	SoftRemoveFile(ctx context.Context, tx transaction.RollbackCommitter, fileUUID string) error

	// HardRemoveFile removes a file record completely from the db.
	HardRemoveFile(ctx context.Context, tx transaction.RollbackCommitter, fileUUID string, rmFileFn func(string) error) error
}
