package models

import (
	"fmt"
	"time"
)

const (
	InternalServerErrorMessage = "Oops! Something wrong happened in our server."
)

type FManError struct {
	Code    int
	Message string
	ErrTime time.Time
}

func (e FManError) Error() string {
	return fmt.Sprintf("Error code: %d. Error message: %s. Time: %s", e.Code, e.Message, e.ErrTime.String())
}
