package authen

import "context"

// Authenticator holds operations on authenticate users.
type Authenticator interface {
	// Register registers a user.
	// Input: username (string), password (string).
	// Output: If registration succeeds, return nil;
	// otherwise, return an error.
	// There are 2 types of errors:
	// - Internal server error happening when something is broken.
	// - Bad username error happening when the username already exists.
	Register(ctx context.Context, username, password string) error

	// Login authenticates a user.
	// Input: username (string), password (string).
	// Output: If registration succeeds,
	// return an auth token and a nil error;
	// otherwise, return an empty string and a non-nil error.
	// There are 2 types of errors:
	// - Internal server error happening when something is broken.
	// - Authenticate error happens when username or password are incorrect.
	Login(ctx context.Context, username, password string) (string, error)
}
