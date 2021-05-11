package uuidUtils

import (
	guuid "github.com/google/uuid"
)

// UUIDGenerateValidator provides an interface for generating a new UUID
// and validating a specific UUID.
type UUIDGenerateValidator interface {
	// NewUUID generates a new unique UUID.
	NewUUID() string

	// ValidateUUID validates a specific UUID.
	ValidateUUID(uuid string) bool
}

// GoogleUUIDGenerator Google UUID generator.
type GoogleUUIDGenerator struct{}

// NewUUID generates a new UUIDv4 (using google/uuid lib).
func (g *GoogleUUIDGenerator) NewUUID() string {
	return guuid.NewString()
}

// ValidateUUID validates a specific UUID (using google/uuid lib).
func (g *GoogleUUIDGenerator) ValidateUUID(uuid string) bool {
	_, err := guuid.Parse(uuid)
	return err == nil
}
