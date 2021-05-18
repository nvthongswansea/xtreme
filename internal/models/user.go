package models

import "time"

// User holds properties of a user.
type User struct {
	UUID      string    `json:"uuid"`
	Username  string    `json:"username"`
	HashPwd   string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserRole struct {
	UserUUID string
}