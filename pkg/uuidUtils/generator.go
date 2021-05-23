package uuidUtils

import (
	guuid "github.com/google/uuid"
)

// UUIDGenerateValidator provides an interface for generating a new UserUUID
// and validating a specific UserUUID.
type UUIDGenerateValidator interface {
	UUIDGenerator

	// ValidateUUID validates a specific UserUUID.
	ValidateUUID(uuid string) bool
}

type UUIDGenerator interface {
	// NewUUID generates a new unique UserUUID.
	NewUUID() string
}

// GoogleUUIDGenerator Google UserUUID generator.
type GoogleUUIDGenerator struct{}

// NewUUID generates a new UUIDv4 (using google/uuid lib).
func (g GoogleUUIDGenerator) NewUUID() string {
	return guuid.NewString()
}

// ValidateUUID validates a specific UserUUID (using google/uuid lib).
func (g GoogleUUIDGenerator) ValidateUUID(uuid string) bool {
	_, err := guuid.Parse(uuid)
	return err == nil
}
