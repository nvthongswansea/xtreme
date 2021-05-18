package local

import (
	"context"
	"io"

	"github.com/nvthongswansea/xtreme/internal/models"
)

// FileManagerService provides an interface for handling files/directories on local storage service.
type FileManagerService interface {
	// GetDirUUIDByPath gets a directory's UUID by a given path.
	GetDirUUIDByPath(ctx context.Context, userUUID, path string) (string, error)

	// CreateNewFile creates a new empty file with a given name.
	// It returns a new file UUID, if success.
	CreateNewFile(ctx context.Context, userUUID, filename, parentDirUUID string) (string, error)

	// UploadFile uploads a single file. The content of the file/data stream
	// should be readable (and closable) via an io.ReadCloser instance.
	// UploadFile returns a new file UUID, if success.
	UploadFile(ctx context.Context, userUUID, filename, parentDirUUID string, contentReader io.Reader) (string, error)

	// CopyFile copies a file to a new location.
	// Note: Beside the new parent dir, the owner
	// of the file could be changed.
	CopyFile(ctx context.Context, userUUID, fileUUID, dstParentDirUUID string) (string, error)

	// RenameFile renames a file.
	RenameFile(ctx context.Context, userUUID, fileUUID, newFileName string) error

	// MoveFile moves a file to a new destination directory.
	MoveFile(ctx context.Context, userUUID, fileUUID, dstParentDirUUID string) error

	// GetFile gets a File object.
	GetFile(ctx context.Context, userUUID, fileUUID string) (models.File, error)

	// DownloadFile gets a file payload. Used for downloading file content.
	// Remember to close the reader to prevent leak.
	DownloadFile(ctx context.Context, userUUID, fileUUID string) (models.FilePayload, error)

	// DownloadFileBatch gets a tmp file payload.
	// It compresses selected files into one single file
	// used for downloading temp file content.
	// Remember to close the tmp file reader to prevent leak, then remove the tmp file.
	DownloadFileBatch(ctx context.Context, userUUID string, fileUUIDs []string) (models.TmpFilePayload, error)

	// SearchByName searches all files/directories based on given filename within a specific path.
	SearchByName(ctx context.Context, userUUID, filename, parentDirUUID string) ([]models.FileMetadata, []models.DirectoryMetadata, error)

	// SoftRemoveFile removes a file (soft).
	// the removed file can be restored in the future.
	SoftRemoveFile(ctx context.Context, userUUID, fileUUID string) error

	// HardRemoveFile removes a file (hard).
	// the removed file can NOT be restored in the future.
	HardRemoveFile(ctx context.Context, userUUID, fileUUID string) error

	// CreateNewDirectory creates a new directory/folder.
	CreateNewDirectory(ctx context.Context, userUUID, dirname, parentDirUUID string) (string, error)

	// GetDirectory gets a directory object.
	GetDirectory(ctx context.Context, userUUID, dirUUID string) (models.Directory, error)

	// GetRootDirectory gets root directory of a user.
	GetRootDirectory(ctx context.Context, userUUID string) (models.Directory, error)

	// CopyDirectory copies a directory/folder to a new location.
	CopyDirectory(ctx context.Context, userUUID, dirUUID, dstParentDirUUID string) (string, error)

	// RenameDirectory renames a directory.
	RenameDirectory(ctx context.Context, userUUID, dirUUID, newDirName string) error

	// MoveDirectory moves a directory/folder to a new location.
	MoveDirectory(ctx context.Context, userUUID, dirUUID, dstParentDirUUID string) (string, error)

	// DownloadDirectory returns a temp file payload.
	// Compress files in selected directory into one single
	// file used for downloading temp file content.
	// Remember to close the tmp file reader to prevent leak, then remove the tmp file.
	DownloadDirectory(ctx context.Context, userUUID, dirUUID string) (models.TmpFilePayload, error)

	// SoftRemoveDir removes a directory and its children (soft).
	// The directory and its children can be restored later.
	SoftRemoveDir(ctx context.Context, userUUID, dirUUID string) error

	// HardRemoveDir removes a directory and its children (hard).
	// The directory and its children can NOT be restored later.
	HardRemoveDir(ctx context.Context, userUUID, dirUUID string) error
}
