package role

import "context"

// Repository defines user role getter methods w.r.t specific
// file/directory.
type Repository interface {
	// GetUserRoleByFile returns role of a user upon the given file.
	// Current accepted roles: "owner", "editor", "viewer".
	GetUserRoleByFile(ctx context.Context, userUUID, fileUUID string) (string, error)

	// GetUserRoleByDirectory returns role of a user upon the given directory.
	// Current accepted roles: "owner", "editor", "viewer".
	GetUserRoleByDirectory(ctx context.Context, userUUID, dirUUID string) (string, error)
}
