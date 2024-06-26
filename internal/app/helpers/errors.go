package helpers

import (
	"errors"
	"fmt"
)

var ErrConflict = errors.New("status 409 conflict")
var ErrIsDeleted = "Short url is deleted"

type ConflictError struct {
	Err      error
	ShortURL string
}

func (ce *ConflictError) Error() string {
	return fmt.Sprintf("Conflict Error. ShortURL already exists: %s, Error: %v", ce.ShortURL, ce.Err)
}

func NewIsDeletedErr(err string) error {
	return fmt.Errorf("%s: %s", err, ErrIsDeleted)
}
