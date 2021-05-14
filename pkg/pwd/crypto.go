package pwd

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	defaultHashCost = 10
)

// HashComparer defines operations on hash and password.
type HashComparer interface {
	// GetPwdHash returns hash string of an input password.
	GetPwdHash(password string) (string, error)

	// CompareHashAndPwd checks if a hash string is
	// the hash of a password.
	// Returns true on success, or false on failure.
	CompareHashAndPwd(password, hash string) bool
}

// BCryptHashComparer is bcrypt implementation of HashComparer interface.
type BCryptHashComparer struct {
	hashCost int
}

// NewBCryptHashComparer returns a new BCryptHashComparer instance.
func NewBCryptHashComparer(hashCost int) BCryptHashComparer {
	return BCryptHashComparer{hashCost}
}

// GetPwdHash returns hash string of an input password using bcrypt.
func (b BCryptHashComparer) GetPwdHash(password string) (string, error) {
	hashCost := defaultHashCost
	if b.hashCost > bcrypt.MinCost {
		hashCost = b.hashCost
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), hashCost) //Hash string by cost of 10
	if err != nil {
		return "", err
	}
	return string(bytes), err
}

// CompareHashAndPwd checks if a hash string is
// the bcrypt hash of a password.
// Returns true on success, or false on failure.
func (b BCryptHashComparer) CompareHashAndPwd(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
