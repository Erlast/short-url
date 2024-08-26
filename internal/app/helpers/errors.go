package helpers

import (
	"errors"
	"fmt"
)

// ErrConflict ошибка конфликта записей.
var ErrConflict = errors.New("status 409 conflict")

// ErrIsDeleted оишбка удаления короткой ссылки.
var ErrIsDeleted = "Short url is deleted"

// ConflictError структура ошибки конфликта коротких ссылок.
type ConflictError struct {
	Err      error
	ShortURL string
}

// Error форматирование вывода ошибки конфликта.
func (ce *ConflictError) Error() string {
	return fmt.Sprintf("Conflict Error. ShortURL already exists: %s, Error: %v", ce.ShortURL, ce.Err)
}

// NewIsDeletedErr форматирование ошибки удаления.
func NewIsDeletedErr(err string) error {
	return fmt.Errorf("%s: %s", err, ErrIsDeleted)
}
