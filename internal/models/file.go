package models

import (
	"time"
)

// File holds properties of a File.
type File struct {
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
	FileSize uint64

	// Time when the file is created.
	CreatedAt time.Time

	// Time of the last file update.
	UpdatedAt time.Time
}

// Directory holds properties of a directory/folder.
type Directory struct {
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
	ListOfFiles []File

	// A list of child-dirs.
	ListOfDirs []Directory
}
