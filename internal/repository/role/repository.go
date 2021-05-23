package role

import (
	"context"
	"github.com/nvthongswansea/xtreme/internal/repository/transaction"
)

// Repository defines user role getter methods w.r.t specific
// file/directory.
type Repository interface {
	// GetUserRoleByFile returns role of a user upon the given file.
	// Current accepted roles: "owner", "editor", "viewer".
	GetUserRoleByFile(ctx context.Context, tx transaction.RollbackCommitter, fileUUID, userUUID string) (string, error)

	// GetUserRoleByDirectory returns role of a user upon the given directory.
	// Current accepted roles: "owner", "editor", "viewer".
	GetUserRoleByDirectory(ctx context.Context, tx transaction.RollbackCommitter, dirUUID, userUUID string) (string, error)
}
