package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrConflict(t *testing.T) {
	// Проверяем, что ErrConflict имеет правильное сообщение
	expectedMessage := "status 409 conflict"
	assert.Equal(t, expectedMessage, ErrConflict.Error())
}

func TestConflictError_Error(t *testing.T) {
	// Создаем экземпляр ConflictError
	ce := &ConflictError{
		Err:      ErrConflict,
		ShortURL: "http://short.url",
	}

	// Проверяем форматирование ошибки
	expectedMessage := "Conflict Error. ShortURL already exists: http://short.url, Error: status 409 conflict"
	assert.Equal(t, expectedMessage, ce.Error())
}

func TestNewIsDeletedErr(t *testing.T) {
	// Создаем ошибку с помощью NewIsDeletedErr
	err := NewIsDeletedErr("Some context")

	// Проверяем форматирование ошибки
	expectedMessage := "Some context: Short url is deleted"
	assert.Equal(t, expectedMessage, err.Error())
}
