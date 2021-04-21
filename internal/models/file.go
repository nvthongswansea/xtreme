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

	// Is the file is a directory.
	IsDir bool

	// Human-readable path of the file.
	Path string

	// Real path of the file, where it is logically stored in the disk.
	RealPath string

	// Parent directory UUID.
	ParentUUID string

	// Time when the file is created.
	CreatedAt time.Time

	// Time of the last file update.
	UpdatedAt time.Time
}
