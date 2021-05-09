package models

type CreateFileRequest struct {
	Filename   string `json:"filename"`
	ParentUUID string `json:"parent_uuid"`
}
