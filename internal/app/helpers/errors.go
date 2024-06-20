package helpers

import (
	"errors"
	"fmt"
)

var ErrConflict = errors.New("status 409 conflict")

type ConflictError struct {
	Err      error
	ShortURL string
}

type IsDeletedError struct {
	Err error
}

func (ce *ConflictError) Error() string {
	return fmt.Sprintf("Conflict Error. ShortURL already exists: %s, Error: %v", ce.ShortURL, ce.Err)
}

func (ide *IsDeletedError) Error() string {
	return fmt.Sprintf("Soft delete detected Error: %v", ide.Err)
}
