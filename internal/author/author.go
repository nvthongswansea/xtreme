package author

// Authorizer holds operations on authorization.
type Authorizer interface {
	// Authorize checks if a "user" has permission to
	// do an "action" to a specific "entity".
	// Input: userUUID (string), action (string), entityUUID (string), entityType (string).
	// Output: If authorization succeeds, return true/false, nil error;
	// otherwise, return false and an error.
	// There are 2 types of errors:
	// - Internal server error happening when something is broken.
	// - Bad input arguments (e.g., invalid action, invalid entityUUID, etc.).
	Authorize(userUUID, action, entityUUID, entityType string) (bool, error)
}