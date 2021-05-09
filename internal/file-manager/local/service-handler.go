package local

import (
	"context"
	"io"

	"github.com/nvthongswansea/xtreme/internal/models"
)

// LocalFManServiceHandler provides an interface for handling files/directories on local storage service.
type LocalFManServiceHandler interface {
	// Get a file/directory's UUID by a given path.
	GetUUIDByPath(ctx context.Context, userUUID, path string) (string, error)

	// Create a new file.
	CreateNewFile(ctx context.Context, userUUID, filename, parentDirUUID string) (string, error)

	// Uploads a single file.
	UploadFile(ctx context.Context, userUUID, filename, parentDirUUID string, fileReadCloser io.ReadCloser) (string, error)

	// Copy a file to a new location.
	CopyFile(ctx context.Context, userUUID, fileUUID, dstParentDirUUID string) (string, error)

	// Rename a file.
	RenameFile(ctx context.Context, userUUID, fileUUID, newFileName string) error

	// Move a file.
	MoveFile(ctx context.Context, userUUID, fileUUID, dstParentDirUUID string) (string, error)

	// Get a file.
	GetFile(ctx context.Context, userUUID, fileUUID string) (models.File, error)

	// Get a file payload. Used for downloading file content. Remember to close the reader to prevent leak.
	DownloadFile(ctx context.Context, userUUID, fileUUID string) (models.FilePayload, error)

	// Compress selected files into one single file. Return a temp file payload. Used for downloading temp file content.
	// Remember to close the tmp file reader to prevent leak, then remove the tmp file.
	DownloadFileBatch(ctx context.Context, userUUID string, fileUUIDs []string) (models.TmpFilePayload, error)

	// Search all files/directories based on given filename within a specific path.
	SearchByName(ctx context.Context, userUUID, filename, parentDirUUID string) ([]models.File, []models.Directory, error)

	// Soft remove a file.
	SoftRemoveFile(ctx context.Context, userUUID, fileUUID string) error

	// Hard remove a file.
	HardRemoveFile(ctx context.Context, userUUID, fileUUID string) error

	// Create a new directory/folder.
	CreateNewDirectory(ctx context.Context, userUUID, dirname, parentDirUUID string) (string, error)

	// Get a directory metadata.
	GetDirectoryMeta(ctx context.Context, userUUID, dirUUID string) (models.Directory, error)

	// Copy a directory/folder to a new location.
	CopyDirectory(ctx context.Context, userUUID, dirUUID, dstParentDirUUID string) (string, error)

	// Rename a directory.
	RenameDirectory(ctx context.Context, userUUID, dirUUID, newDirName string) error

	// Move a directory/folder to a new location.
	MoveDirectory(ctx context.Context, userUUID, dirUUID, dstParentDirUUID string) (string, error)

	// Compress files in selected directory into one single file. Return a temp file payload. Used for downloading temp file content.
	// Remember to close the tmp file reader to prevent leak, then remove the tmp file.
	DownloadDirectory(ctx context.Context, userUUID, dirUUID string) (models.TmpFilePayload, error)

	// Soft remove a directory.
	SoftRemoveDir(ctx context.Context, userUUID, dirUUID string) error

	// Hard remove a directory.
	HardRemoveDir(ctx context.Context, userUUID, dirUUID string) error
}
