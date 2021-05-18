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
	// Conditions for success:
	// - User has permission to create a new file in
	// the given directory.
	// - The filename is not taken by any child-files/dirs in
	// the given directory.
	CreateNewFile(ctx context.Context, userUUID, filename, parentDirUUID string) (string, error)

	// UploadFile uploads a single file. The content of the file/data stream
	// should be readable (and closable) via an io.ReadCloser instance.
	// UploadFile returns a new file UUID, if success.
	// Conditions for success:
	// - User has permission to create a new file in
	// the given directory.
	// - The filename is not taken by any child-files/dirs in
	// the given directory.
	// - The data stream in io.ReadCloser is valid.
	UploadFile(ctx context.Context, userUUID, filename, parentDirUUID string, contentReader io.Reader) (string, error)

	// CopyFile copies a file to a new location.
	// Note: Beside the new parent dir, the owner
	// of the file could be changed.
	CopyFile(ctx context.Context, userUUID, fileUUID, dstParentDirUUID string) (string, error)

	// RenameFile renames a file.
	// This function succeeds, if:
	// - The file exists (within user storage space context).
	// - The new name doesn't present in the current directory.
	RenameFile(ctx context.Context, userUUID, fileUUID, newFileName string) error

	// Move a file.
	MoveFile(ctx context.Context, userUUID, fileUUID, dstParentDirUUID string) error

	// Get a file.
	GetFile(ctx context.Context, userUUID, fileUUID string) (models.File, error)

	// Get a file payload. Used for downloading file content. Remember to close the reader to prevent leak.
	DownloadFile(ctx context.Context, userUUID, fileUUID string) (models.FilePayload, error)

	// Compress selected files into one single file. Return a temp file payload. Used for downloading temp file content.
	// Remember to close the tmp file reader to prevent leak, then remove the tmp file.
	DownloadFileBatch(ctx context.Context, userUUID string, fileUUIDs []string) (models.TmpFilePayload, error)

	// Search all files/directories based on given filename within a specific path.
	SearchByName(ctx context.Context, userUUID, filename, parentDirUUID string) ([]models.FileMetadata, []models.DirectoryMetadata, error)

	// Soft remove a file.
	SoftRemoveFile(ctx context.Context, userUUID, fileUUID string) error

	// Hard remove a file.
	HardRemoveFile(ctx context.Context, userUUID, fileUUID string) error

	// Create a new directory/folder.
	CreateNewDirectory(ctx context.Context, userUUID, dirname, parentDirUUID string) (string, error)

	// Get a directory.
	GetDirectory(ctx context.Context, userUUID, dirUUID string) (models.Directory, error)

	// Get root directory.
	GetRootDirectory(ctx context.Context, userUUID string) (models.Directory, error)

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
