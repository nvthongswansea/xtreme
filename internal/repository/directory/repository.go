package directory

import (
	"context"
	"github.com/nvthongswansea/xtreme/internal/models"
)

// Repository holds all operations on directories.
type Repository interface {
	Inserter
	Reader
	Updater
	Remover
}

// Inserter holds inserting operations on directories.
type Inserter interface {
	// InsertDirectoryMetadata inserts a directory/folder metadata to db.
	InsertDirectoryMetadata(ctx context.Context, newDir models.DirectoryMetadata) error
}

// Reader holds reading operations on directories.
type Reader interface {
	// GetDirectory gets a directory/folder record from the db with a given UUID.
	GetDirectory(ctx context.Context, dirUUID string) (models.Directory, error)

	// GetDirMetadata gets directory/folder metadata from the db with a given UUID.
	GetDirMetadata(ctx context.Context, dirUUID string) (models.DirectoryMetadata, error)

	// GetDirMetadataListByName gets all directories' metadata named after the given dirname.
	GetDirMetadataListByName(ctx context.Context, dirname, parentDirUUID string) ([]models.DirectoryMetadata, error)

	// GetRootDirectoryByUserUUID gets a user's root directory from the db.
	GetRootDirectoryByUserUUID(ctx context.Context, userUUID string) (models.Directory, error)

	// GetDirUUIDByPath gets a file/directory's UUID by a given path.
	GetDirUUIDByPath(ctx context.Context, userUUID, path string) (string, error)

	// GetDirectChildDirUUIDList get direct child-directories' UUIDs
	// of a specific directory.
	GetDirectChildDirUUIDList(ctx context.Context, dirUUID string) ([]string, error)

	// IsDirNameExist checks if a specific dirname exists in a specific directory.
	IsDirNameExist(ctx context.Context, name, parentDirUUID string) (bool, error)
}

// Updater holds updating operations on directories.
type Updater interface {
	// UpdateDirname updates name of a directory in the db.
	UpdateDirname(ctx context.Context, dirUUID, newDirname string) error

	// UpdateParentDirUUID updates the parent dirUUID of a directory.
	UpdateParentDirUUID(ctx context.Context, dirUUID, newParentDirUUID string) error
}

// Remover holds removing operations on directories.
type Remover interface {
	// SoftRemoveDir flags a directory/folder record as deleted file.
	// e.g. set `is_deleted` field to true.
	SoftRemoveDir(ctx context.Context, dirUUID string) error

	// HardRemoveDir removes a directory/folder record completely from the db.
	HardRemoveDir(ctx context.Context, dirUUID string, fileRmCallback func(string) error) error
}
