package database

import (
	"context"
	"github.com/nvthongswansea/xtreme/internal/models"
)

// Repository holds all operations to deal with the database.
type Repository interface {
	DBEntityInserter
	DBEntityGetter
	DBEntitySearcher
	DBEntityUpdater
	DBEntityRemover
	DBEntityExistenceChecker
}

// LocalFManRepository holds all operations on the db,
// which are used by local storage's file manager.
type LocalFManRepository interface {
	DBEntityMetadataInserter
	DBEntityGetter
	DBEntitySearcher
	DBEntityMetadataUpdater
	DBEntityRemover
	DBEntityExistenceChecker
}

// DBEntityInserter holds inserting operations on db.
type DBEntityInserter interface {
	DBEntityMetadataInserter
}

// DBEntityMetadataInserter holds metadata inserting operations on db.
type DBEntityMetadataInserter interface {
	// InsertFileMetadata inserts a new file metadata to db.
	InsertFileMetadata(ctx context.Context, newFile models.FileMetadata) error

	// InsertDirectoryMetadata inserts a directory/folder metadata to db.
	InsertDirectoryMetadata(ctx context.Context, newDir models.DirectoryMetadata) error
}

// DBEntityGetter holds retrieval operations on db.
type DBEntityGetter interface {
	// GetFile gets a file record from the db with a given UUID.
	GetFile(ctx context.Context, userUUID, fileUUID string) (models.File, error)

	// GetUUIDByPath gets a file/directory's UUID by a given path.
	GetUUIDByPath(ctx context.Context, userUUID, path string) (string, error)

	// GetDirectory gets a directory/folder record from the db with a given UUID.
	GetDirectory(ctx context.Context, userUUID, dirUUID string) (models.Directory, error)

	DBEntityMetadataGetter
}

// DBEntityMetadataGetter holds metadata retrieval operations on db.
type DBEntityMetadataGetter interface {
	// GetFileMetadata gets file metadata from the db with a given UUID.
	GetFileMetadata(ctx context.Context, userUUID, fileUUID string) (models.FileMetadata, error)

	// GetFileMetadataBatch gets multiple files' metadata from the db with given UUIDs.
	GetFileMetadataBatch(ctx context.Context, userUUID string, fileUUIDs []string) ([]models.FileMetadata, error)

	// GetDirMetadata gets directory/folder metadata from the db with a given UUID.
	GetDirMetadata(ctx context.Context, userUUID, dirUUID string) (models.DirectoryMetadata, error)
}

// DBEntityUpdater holds updating operations on db.
type DBEntityUpdater interface {
	DBEntityMetadataUpdater
}

// DBEntityMetadataUpdater holds metadata updating operations on db
type DBEntityMetadataUpdater interface {
	// UpdateFileMetadata updates a file metadata in the db.
	UpdateFileMetadata(ctx context.Context, newMetadata models.FileMetadata) error

	// UpdateDirMetadata updates a directory/folder metadata in the db.
	UpdateDirMetadata(ctx context.Context, newMetadata models.DirectoryMetadata) error
}

// DBEntityRemover holds removing operations on db.
type DBEntityRemover interface {
	// SoftRemoveFileRecord flags a file record as deleted file.
	// e.g. set `is_deleted` field to true.
	SoftRemoveFileRecord(ctx context.Context, userUUID, fileUUID string) error

	// HardRemoveFileRecord removes a file record completely from the db.
	HardRemoveFileRecord(ctx context.Context, userUUID, fileUUID string) error

	// SoftRemoveDirRecord flags a directory/folder record as deleted file.
	// e.g. set `is_deleted` field to true.
	SoftRemoveDirRecord(ctx context.Context, userUUID, dirUUID string) error

	// HardRemoveDirRecord removes a directory/folder record completely from the db.
	HardRemoveDirRecord(ctx context.Context, userUUID, dirUUID string) error
}

// DBEntityExistenceChecker holds validating operations on db.
type DBEntityExistenceChecker interface {
	// IsNameExist checks if a specific file/dir's name exists in a specific directory.
	IsNameExist(ctx context.Context, userUUID, name, parentDirUUID string) (bool, error)

	// IsDirExist checks if a parent directory UUID exists.
	IsDirExist(ctx context.Context, userUUID, parentDirUUID string) (bool, error)
}

// DBEntitySearcher holds searching operations on db.
type DBEntitySearcher interface {
	// Search all files/directories based on given filename within a specific path.
	SearchByName(ctx context.Context, userUUID, filename, parentDirUUID string) ([]models.File, []models.Directory, error)
}
