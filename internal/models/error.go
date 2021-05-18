package models

import (
	"fmt"
)

const (
	InternalServerErrorMessage = "Oops! Something wrong happened in our server."
)

const (
	InternalServerErrorCode = iota
	BadInputErrorCode
	ForbiddenOperationErrorCode
	NotFoundErrorCode
)

type XtremeError struct {
	Code    int    `json:"-"`
	Message string `json:"message"`
}

func (e XtremeError) Error() string {
	return fmt.Sprintf("Error code: %d. Error message: %s.", e.Code, e.Message)
}
