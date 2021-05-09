package models

import "time"

// Directory holds metadata and content of a directory/folder.
type Directory struct {
	Metadata DirectoryMetadata `json:"metadata"`
	Content DirectoryContent `json:"content"`
}

// DirectoryMetadata holds metadata of a directory/folder.
type DirectoryMetadata struct {
	// UUID of the directory.
	UUID string `json:"uuid"`

	// Name of the directory.
	Dirname string `json:"dirname"`

	// Human-readable path of the directory.
	Path string `json:"path"`

	// Parent directory UUID.
	ParentUUID string `json:"parent_uuid"`

	// Directory owner's userUUID.
	OwnerUUID string `json:"owner_uuid"`

	// Time when the directory is created.
	CreatedAt time.Time `json:"created_at"`

	// Time of the last directory update.
	UpdatedAt time.Time `json:"updated_at"`
}

// DirectoryContent holds content of a directory/folder.
type DirectoryContent struct {
	// A list of direct child-files metadata.
	ListOfFiles []FileMetadata

	// A list of direct child-dirs metadata.
	ListOfDirs []DirectoryMetadata
}