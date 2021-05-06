package fman

import "github.com/nvthongswansea/xtreme/internal/models"

// FManFileDBRepo provides an interface for operations on file in the database.
type FManFileDBRepo interface {
	// InsertFileRecord inserts a file record to db.
	InsertFileRecord(UUID, filename, parentUUID, realPath string, fileSize int64) error

	// ReadFileRecord reads a file record from the db with a given UUID.
	ReadFileRecord(UUID string) (models.File, error)

	// UpdateFileRecord updates a file record in the db.
	UpdateFileRecord(filename, parentUUID string) error

	// SoftRemoveFileRecord flags a file record as deleted file.
	// e.g. set `is_deleted` field to true.
	SoftRemoveFileRecord(UUID string) error

	// HardRemoveFileRecord removes a file record completely from the db.
	HardRemoveFileRecord(UUID string) error
}

// FManDirDBRepo provides an interface for operations on directory/folder in the database.
type FManDirDBRepo interface {
	// InsertDirRecord inserts a directory/folder record to db.
	InsertDirRecord(UUID, dirname, parentUUID string) error

	// ReadDirRecord reads a directory/folder record from the db with a given UUID.
	ReadDirRecord(UUID string) (models.Directory, error)

	// UpdateDirRecord updates a directory/folder record in the db.
	UpdateDirRecord(filename, parentUUID string) error

	// SoftRemoveDirRecord flags a directory/folder record as deleted file.
	// e.g. set `is_deleted` field to true.
	SoftRemoveDirRecord(UUID string) error

	// HardRemoveDirRecord removes a directory/folder record completely from the db.
	HardRemoveDirRecord(UUID string) error
}

// FManValidateDBRepo provides an interface for operations on data validation via db.
type FManValidateDBRepo interface {
	// IsNameExist checks if a specific file/dir's name exists in a specific path.
	IsNameExist(filename, parentUUID string) (bool, error)

	// IsParentUUIDExist checks if a parent UUID exists.
	IsParentUUIDExist(parentUUID string) (bool, error)
}
