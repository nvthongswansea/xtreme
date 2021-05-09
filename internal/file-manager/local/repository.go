package local

import (
	"context"

	"github.com/nvthongswansea/xtreme/internal/models"
)

// LocalFManDBRepo provides an interface for operations on files/directories in the database.
type LocalFManDBRepo interface {
	// InsertFileRecord inserts a file record to db.
	InsertFileRecord(ctx context.Context, userUUID, fileUUID, filename, parentUUID, realPath string, fileSize int64) error

	// GetFileRecord gets a file record from the db with a given UUID.
	GetFile(ctx context.Context, userUUID, fileUUID string) (models.File, error)

	// GetFileRecordBatch gets multiple file records from the db with given UUIDs.
	GetFileBatch(ctx context.Context, userUUID string, fileUUIDs []string) ([]models.File, error)

	// Get a file/directory's UUID by a given path.
	GetUUIDByPath(ctx context.Context, userUUID, path string) (string, error)

	// Search all files/directories based on given filename within a specific path.
	SearchByName(ctx context.Context, userUUID, filename, parentDirUUID string) ([]models.File, []models.Directory, error)

	// UpdateFileRecord updates a file record in the db.
	UpdateFileRecord(ctx context.Context, userUUID, newFileName, parentUUID string) error

	// SoftRemoveFileRecord flags a file record as deleted file.
	// e.g. set `is_deleted` field to true.
	SoftRemoveFileRecord(ctx context.Context, userUUID, fileUUID string) error

	// HardRemoveFileRecord removes a file record completely from the db.
	HardRemoveFileRecord(ctx context.Context, userUUID, fileUUID string) error

	// InsertDirRecord inserts a directory/folder record to db.
	InsertDirRecord(ctx context.Context, userUUID, dirUUID, dirname, parentUUID string) error

	// GetDirRecord gets a directory/folder record from the db with a given UUID.
	GetDirectory(ctx context.Context, userUUID, dirUUID string) (models.Directory, error)

	// UpdateDirRecord updates a directory/folder record in the db.
	UpdateDirRecord(ctx context.Context, userUUID, newDirName, parentUUID string) error

	// SoftRemoveDirRecord flags a directory/folder record as deleted file.
	// e.g. set `is_deleted` field to true.
	SoftRemoveDirRecord(ctx context.Context, userUUID, dirUUID string) error

	// HardRemoveDirRecord removes a directory/folder record completely from the db.
	HardRemoveDirRecord(ctx context.Context, userUUID, dirUUID string) error

	// IsNameExist checks if a specific file/dir's name exists in a specific directory.
	IsNameExist(ctx context.Context, userUUID, name, parentDirUUID string) (bool, error)

	// IsDirExist checks if a parent directory UUID exists.
	IsDirExist(ctx context.Context, userUUID, parentDirUUID string) (bool, error)
}
