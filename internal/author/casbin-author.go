package author

import (
	"context"
	"github.com/casbin/casbin/v2"
	"github.com/nvthongswansea/xtreme/internal/repository/role"
	"github.com/nvthongswansea/xtreme/internal/repository/transaction"
)

// CasbinAuthorizer is a casbin implementation of
// the interface Authorizer.
type CasbinAuthorizer struct {
	enforcer *casbin.Enforcer
	roleRepo role.Repository
}

// NewCasbinAuthorizer create a new casbin enforcer from model filepath
// and policy filepath.
func NewCasbinAuthorizer(modelPath, policyPath string, roleRepo role.Repository) (*CasbinAuthorizer, error) {
	enforcer, err := casbin.NewEnforcer(modelPath, policyPath)
	return &CasbinAuthorizer{enforcer, roleRepo}, err
}

// AuthorizeActionsOnFile checks if a user has a right to perform some actions on a certain file.
func (c *CasbinAuthorizer) AuthorizeActionsOnFile(ctx context.Context, tx transaction.RollbackCommitter, userUUID, fileUUID string, actions ...fileAction) (bool, error) {
	// Get user role.
	role, err := c.roleRepo.GetUserRoleByFile(ctx, tx, fileUUID, userUUID)
	if err != nil {
		return false, err
	}
	for _, action := range actions {
		isAllowed, err := c.enforcer.Enforce(role, "file", action)
		if err != nil {
			return false, err
		}
		if !isAllowed {
			return false, nil
		}
	}
	return true, nil
}

// AuthorizeActionsOnDir checks if a user has a right to perform some actions on a certain directory.
func (c *CasbinAuthorizer) AuthorizeActionsOnDir(ctx context.Context, tx transaction.RollbackCommitter, userUUID, dirUUID string, actions ...dirAction) (bool, error) {
	// Get user role.
	role, err := c.roleRepo.GetUserRoleByDirectory(ctx, tx, dirUUID, userUUID)
	if err != nil {
		return false, err
	}
	for _, action := range actions {
		isAllowed, err := c.enforcer.Enforce(role, "directory", action)
		if err != nil {
			return false, err
		}
		if !isAllowed {
			return false, nil
		}
	}
	return true, nil
}
