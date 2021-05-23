package directory

import (
	"context"
	"github.com/nvthongswansea/xtreme/internal/models"
	"github.com/nvthongswansea/xtreme/internal/repository/transaction"
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
	// InsertDirectory inserts a directory metadata to db, returns
	// inserted directory's UserUUID (if success).
	InsertDirectory(ctx context.Context, tx transaction.RollbackCommitter, newDir models.Directory) (string, error)

	// InsertRootDirectory inserts a new user's root directory to the db.
	InsertRootDirectory(ctx context.Context, tx transaction.RollbackCommitter, userUUID string) (string, error)
}

// Reader holds reading operations on directories.
type Reader interface {
	// GetDirectory gets a directory/folder record from the db with a given UserUUID.
	GetDirectory(ctx context.Context, tx transaction.RollbackCommitter, dirUUID string) (models.Directory, error)

	// GetDirMetadata gets directory/folder metadata from the db with a given UserUUID.
	GetDirMetadata(ctx context.Context, tx transaction.RollbackCommitter, dirUUID string) (models.DirectoryMetadata, error)

	// GetDirMetadataListByName gets all directories' metadata named after the given dirname.
	GetDirMetadataListByName(ctx context.Context, tx transaction.RollbackCommitter, parentDirUUID, dirname string) ([]models.DirectoryMetadata, error)

	// GetRootDirectoryByUserUUID gets a user's root directory from the db.
	GetRootDirectoryByUserUUID(ctx context.Context, tx transaction.RollbackCommitter, userUUID string) (models.Directory, error)

	// GetDirUUIDByPath gets a file/directory's UserUUID by a given path.
	GetDirUUIDByPath(ctx context.Context, tx transaction.RollbackCommitter, path, userUUID string) (string, error)

	// GetDirectChildDirUUIDList get direct child-directories' UUIDs
	// of a specific directory.
	GetDirectChildDirUUIDList(ctx context.Context, tx transaction.RollbackCommitter, dirUUID string) ([]string, error)

	// IsDirNameExist checks if a specific dirname exists in a specific directory.
	IsDirNameExist(ctx context.Context, tx transaction.RollbackCommitter, parentDirUUID, name string) (bool, error)
}

// Updater holds updating operations on directories.
type Updater interface {
	// UpdateDirname updates name of a directory in the db.
	UpdateDirname(ctx context.Context, tx transaction.RollbackCommitter, newDirname, dirUUID string) error

	// UpdateParentDirUUID updates the parent dirUUID of a directory.
	UpdateParentDirUUID(ctx context.Context, tx transaction.RollbackCommitter, newParentDirUUID, dirUUID string) error
}

// Remover holds removing operations on directories.
type Remover interface {
	// SoftRemoveDir flags a directory/folder record as deleted file.
	// e.g. set `is_deleted` field to true.
	SoftRemoveDir(ctx context.Context, tx transaction.RollbackCommitter, dirUUID string) error

	// HardRemoveDir removes a directory/folder record completely from the db.
	HardRemoveDir(ctx context.Context, tx transaction.RollbackCommitter, dirUUID string) error
}
