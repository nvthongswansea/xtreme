package models

type CreateFileDirRequest struct {
	Name          string `json:"name"`
	ParentDirUUID string `json:"parent_dir_uuid"`
}

type CopyMvRequest struct {
	DstDirUUID string `json:"dst_dir_uuid"`
}

type RenameRequest struct {
	NewName string `json:"new_name"`
}

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthenUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
