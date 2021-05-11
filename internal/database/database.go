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
	DBEntityValidator
}

// LocalFManRepository holds all operations on the db,
// which are used by local storage's file manager.
type LocalFManRepository interface {
	DBEntityMetadataInserter
	DBEntityGetter
	DBEntitySearcher
	DBEntityMetadataUpdater
	DBEntityRemover
	DBEntityValidator
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
	GetFile(ctx context.Context, fileUUID string) (models.File, error)

	// GetDirectory gets a directory/folder record from the db with a given UUID.
	GetDirectory(ctx context.Context, dirUUID string) (models.Directory, error)

	// GetRootDirectory gets a user's root directory from the db.
	GetRootDirectory(ctx context.Context, userUUID string) (models.Directory, error)

	// GetUUIDByPath gets a file/directory's UUID by a given path.
	GetUUIDByPath(ctx context.Context, rootDirUUID, path string) (string, bool, error)

	// GetDirectChildDirUUIDList get direct child-directories' UUIDs
	// of a specific directory.
	GetDirectChildDirUUIDList(ctx context.Context, dirUUID string) ([]string, error)

	DBEntityMetadataGetter
}

// DBEntityMetadataGetter holds metadata retrieval operations on db.
type DBEntityMetadataGetter interface {
	// GetFileMetadata gets file metadata from the db with a given UUID.
	GetFileMetadata(ctx context.Context, fileUUID string) (models.FileMetadata, error)

	// GetFileMetadataBatch gets multiple files' metadata from the db with given UUIDs.
	GetFileMetadataBatch(ctx context.Context, fileUUIDs []string) ([]models.FileMetadata, error)

	// GetChildFileMetadataList gets all child-files' metadata in a given directory.
	GetChildFileMetadataList(ctx context.Context, dirUUID string) ([]models.FileMetadata, error)

	// GetRootDirMetadata gets a user's root directory metadata from the db.
	GetRootDirMetadata(ctx context.Context, userUUID string) (models.DirectoryMetadata, error)

	// GetDirMetadata gets directory/folder metadata from the db with a given UUID.
	GetDirMetadata(ctx context.Context, dirUUID string) (models.DirectoryMetadata, error)
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
	// SoftRemoveFile flags a file record as deleted file.
	// e.g. set `is_deleted` field to true.
	SoftRemoveFile(ctx context.Context, fileUUID string) error

	// HardRemoveFile removes a file record completely from the db.
	HardRemoveFile(ctx context.Context, fileUUID string, fileRmCallback func(string) error) error

	// SoftRemoveDir flags a directory/folder record as deleted file.
	// e.g. set `is_deleted` field to true.
	SoftRemoveDir(ctx context.Context, dirUUID string) error

	// HardRemoveDir removes a directory/folder record completely from the db.
	HardRemoveDir(ctx context.Context, dirUUID string, fileRmCallback func(string) error) error
}

// DBEntityValidator holds validating operations on db.
type DBEntityValidator interface {
	// IsNameExist checks if a specific file/dir's name exists in a specific directory.
	IsNameExist(ctx context.Context, name, parentDirUUID string) (bool, error)

	// IsUserUUIDExist checks if a user UUID exists.
	IsUserUUIDExist(ctx context.Context, userUUID string) (bool, error)

	// IsFileExist checks if a file UUID exists.
	IsFileExist(ctx context.Context, userUUID, fileUUID string) (bool, error)

	// IsDirExist checks if a parent directory UUID exists.
	IsDirExist(ctx context.Context, userUUID, parentDirUUID string) (bool, error)

	// IsRootDir checks if the given directory is a root directory.
	IsRootDir(ctx context.Context, userUUID, dirUUID string) (bool, error)
}

// DBEntitySearcher holds searching operations on db.
type DBEntitySearcher interface {
	// Search all files/directories based on given filename within a specific path.
	SearchByName(ctx context.Context, userUUID, filename, parentDirUUID string) ([]models.File, []models.Directory, error)
}
