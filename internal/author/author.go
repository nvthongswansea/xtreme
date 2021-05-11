package author

import "context"

type fileAction string
const (
	ViewFileAction fileAction = "view"
	UpdateFileAction fileAction = "update"
	CopyFileAction fileAction = "copy"
	RemoveFileAction fileAction = "remove"
)

type dirAction string
const (
	ViewDirAction dirAction = "view"
	UploadToDirAction dirAction = "uploadTo"
	UpdateDirAction dirAction = "update"
	CopyDirAction dirAction = "copy"
	RemoveDirAction dirAction = "remove"
)

// Authorizer holds operations on authorization.
type Authorizer interface {
	// AuthorizeActionsOnFile checks if a "user" has permission to
	// do a series of "actions" to a specific file.
	// Input: userUUID (string), action (fileAction), fileUUID (string).
	// Output: If authorization succeeds, return true/false, nil error;
	// otherwise, return false and an error.
	// Error is an internal server error.
	AuthorizeActionsOnFile(ctx context.Context, userUUID, fileUUID string, actions ...fileAction) (bool, error)

	// AuthorizeActionsOnDir checks if a "user" has permission to
	// do series of "actions" to a specific directory.
	// Input: userUUID (string), action (dirAction), dirUUID (string).
	// Output: If authorization succeeds, return true/false, nil error;
	// otherwise, return false and an error.
	// Error is an internal server error.
	//
	// **Note**: Be careful about the "root" directory of a user.
	AuthorizeActionsOnDir(ctx context.Context, userUUID, dirUUID string, actions ...dirAction) (bool, error)
}
