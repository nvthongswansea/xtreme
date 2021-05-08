package models

import (
	"io"
	"os"
	"time"
)

// FileMetadata holds metadata of a File.
type FileMetadata struct {
	// UUID of the file.
	UUID string

	// Name of the file.
	Filename string

	// Human-readable path of the file.
	Path string

	// Real path of the file, where it is logically stored in the disk.
	RealPath string

	// Parent directory UUID.
	ParentUUID string

	// Size of the file.
	FileSize int64

	// Time when the file is created.
	CreatedAt time.Time

	// Time of the last file update.
	UpdatedAt time.Time
}

// DirectoryMetadata holds metadata of a directory/folder.
type DirectoryMetadata struct {
	// UUID of the directory.
	UUID string

	// Name of the directory.
	Dirname string

	// Human-readable path of the directory.
	Path string

	// Parent directory UUID.
	ParentUUID string

	// Time when the directory is created.
	CreatedAt time.Time

	// Time of the last directory update.
	UpdatedAt time.Time

	// A list of child-files.
	ListOfFiles []FileMetadata

	// A list of child-dirs.
	ListOfDirs []DirectoryMetadata
}

// FilePayload holds content reader of a file.
type FilePayload struct {
	Filename      string
	ContentLength int64
	ReadCloser    io.ReadCloser
}

// TmpFilePayload holds content reader of a temp file.
// With *os.File, the temp file can be removed after use.
type TmpFilePayload struct {
	Filename      string
	ContentLength int64
	File          *os.File
}
