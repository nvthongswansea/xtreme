package uuidUtils

import (
	guuid "github.com/google/uuid"
)

// UUIDGenerator provides an interface for generating a new UUID.
type UUIDGenerator interface {
	// NewUUID generates a new unique UUID.
	NewUUID() string
}

// GoogleUUIDGenerator Google UUID generator.
type GoogleUUIDGenerator struct{}

// NewUUID generates a new UUIDv4 (using google/uuid lib).
func (g *GoogleUUIDGenerator) NewUUID() string {
	return guuid.NewString()
}
