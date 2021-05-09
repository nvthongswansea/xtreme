package models

import (
	"io"
	"os"
	"time"
)

// File holds metadata and description of a file.
type File struct {
	Metadata FileMetadata`json:"metadata"`
	Description FileDescription `json:"description"`
}

// FileMetadata holds metadata of a File.
type FileMetadata struct {
	// UUID of the file.
	UUID string `json:"uuid"`

	// Name of the file.
	Filename string `json:"filename"`

	// MIME type.
	MIMEType string `json:"mime_type"`

	// Human-readable path of the file.
	Path string `json:"path"`

	// Real absolute path of the file, where it is logically stored.
	AbsPathOnDisk string `json:"-"`

	// Parent directory UUID.
	ParentUUID string `json:"parent_uuid"`

	// Size of the file.
	Size int64 `json:"size"`

	// File owner's userUUID.
	OwnerUUID string `json:"owner_uuid"`

	// Time when the file is created.
	CreatedAt time.Time `json:"created_at"`

	// Time of the last file update.
	UpdatedAt time.Time `json:"updated_at"`
}

// FileDescription describes the file in short form.
type FileDescription struct {
	// Description in text form.
	DescriptionText string `json:"text"`

	// Thumbnail image of the file (base64).
	ThumbnailBase64 string `json:"thumbnail_base_64"`
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
