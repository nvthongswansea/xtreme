package author

import (
	"context"
	"github.com/casbin/casbin/v2"
	"github.com/nvthongswansea/xtreme/internal/database"
)

// CasbinAuthorizer is a casbin implementation of
// the interface Authorizer.
type CasbinAuthorizer struct {
	enforcer *casbin.Enforcer
	roleGetter database.UserRoleByEntityGetter
}

// NewCasbinAuthorizer create a new casbin enforcer from model filepath
// and policy filepath.
func NewCasbinAuthorizer(modelPath, policyPath string, roleGetter database.UserRoleByEntityGetter) (*CasbinAuthorizer, error) {
	enforcer, err := casbin.NewEnforcer(modelPath, policyPath)
	return &CasbinAuthorizer{enforcer, roleGetter}, err
}

// AuthorizeActionsOnFile checks if a user has a right to perform some actions on a certain file.
func (c *CasbinAuthorizer) AuthorizeActionsOnFile(ctx context.Context, userUUID, fileUUID string, actions ...fileAction) (bool, error) {
	// Get user role.
	role, err := c.roleGetter.GetUserRoleByFile(ctx, userUUID, fileUUID)
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
func (c *CasbinAuthorizer) AuthorizeActionsOnDir(ctx context.Context, userUUID, dirUUID string, actions ...dirAction) (bool, error) {
	// Get user role.
	role, err := c.roleGetter.GetUserRoleByDirectory(ctx, userUUID, dirUUID)
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
