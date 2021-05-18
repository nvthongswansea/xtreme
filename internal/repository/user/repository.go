package user

import (
	"context"
	"github.com/nvthongswansea/xtreme/internal/models"
)

// Repository defines operations on user entity.
type Repository interface {
	Inserter
	Reader
}

// Inserter holds inserting operations on user entity.
type Inserter interface {
	// InsertNewUser inserts a new user based on given username and password hash.
	// Returns a new user object, if success.
	InsertNewUser(ctx context.Context, username, pswHash string) error
}

// Reader holds reading operations on user entity.
type Reader interface {
	// GetUserByUsername get an user object based on given username.
	// Returns a retrieved user object, if success.
	GetUserByUsername(ctx context.Context, username string) (models.User, error)

	// IsUsernameExist checks if the username exists.
	IsUsernameExist(ctx context.Context, username string) (bool, error)

	// IsUsernameUsernameExist checks if the (userUUID, username) exists.
	IsUsernameUsernameExist(ctx context.Context, userUUID, username string) (bool, error)
}