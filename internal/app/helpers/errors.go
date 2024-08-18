package helpers

import (
	"errors"
	"fmt"
)

var ErrConflict = errors.New("status 409 conflict") // ErrConflict ошибка конфликта записей
var ErrIsDeleted = "Short url is deleted"           // ErrIsDeleted оишбка удаления короткой ссылки

// ConflictError структура ошибки конфликта коротких ссылок
type ConflictError struct {
	Err      error
	ShortURL string
}

// Error форматирование вывода ошибки конфликта
func (ce *ConflictError) Error() string {
	return fmt.Sprintf("Conflict Error. ShortURL already exists: %s, Error: %v", ce.ShortURL, ce.Err)
}

// NewIsDeletedErr форматирование ошибки удаления
func NewIsDeletedErr(err string) error {
	return fmt.Errorf("%s: %s", err, ErrIsDeleted)
}
