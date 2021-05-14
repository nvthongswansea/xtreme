package database

import (
	"context"
	"github.com/nvthongswansea/xtreme/internal/models"
)

// LocalFManRepository holds all operations on the db,
// which are used by local storage's file manager.
type LocalFManRepository interface {
	// InsertFileMetadata inserts a new file metadata to db.
	InsertFileMetadata(ctx context.Context, newFile models.FileMetadata) error

	// InsertDirectoryMetadata inserts a directory/folder metadata to db.
	InsertDirectoryMetadata(ctx context.Context, newDir models.DirectoryMetadata) error

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

	// UpdateFileMetadata updates a file metadata in the db.
	UpdateFileMetadata(ctx context.Context, newMetadata models.FileMetadata) error

	// UpdateDirMetadata updates a directory/folder metadata in the db.
	UpdateDirMetadata(ctx context.Context, newMetadata models.DirectoryMetadata) error

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

	// IsNameExist checks if a specific file/dir's name exists in a specific directory.
	IsNameExist(ctx context.Context, name, parentDirUUID string) (bool, error)

	// SearchByName searches all files/directories based on given filename within a specific path.
	SearchByName(ctx context.Context, userUUID, filename, parentDirUUID string) ([]models.File, []models.Directory, error)
}

// UserRoleByEntityGetter defines user role getter methods w.r.t specific
// file/directory.
type UserRoleByEntityGetter interface {
	// GetUserRoleByFile returns role of a user upon the given file.
	// Current accepted roles: "owner", "editor", "viewer".
	GetUserRoleByFile(ctx context.Context, userUUID, fileUUID string) (string, error)

	// GetUserRoleByDirectory returns role of a user upon the given directory.
	// Current accepted roles: "owner", "editor", "viewer".
	GetUserRoleByDirectory(ctx context.Context, userUUID, dirUUID string) (string, error)
}

// UserCreateGetter defines create/get operations on user entity.
type UserCreateGetter interface {
	// CreateNewUser creates a new user based on given username and password hash.
	// Returns a new user object, if success.
	CreateNewUser(ctx context.Context, username, pswHash string) error

	// GetUserByUsername get an user object based on given username.
	// Returns a retrieved user object, if success.
	GetUserByUsername(ctx context.Context, username string) (models.User, error)

	// IsUsernameExist checks if the username exists.
	IsUsernameExist(ctx context.Context, username string) (bool, error)

	// IsUsernameUsernameExist checks if the (userUUID, username) exists.
	IsUsernameUsernameExist(ctx context.Context, userUUID, username string) (bool, error)
}
