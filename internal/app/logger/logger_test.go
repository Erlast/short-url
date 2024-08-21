package logger

import (
	"errors"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	t.Run("Valid log level", func(t *testing.T) {
		logger, err := NewLogger("debug")
		assert.NoError(t, err)
		assert.NotNil(t, logger)

		logger.Debug("This is a debug message")

		if err := logger.Sync(); err != nil && !errors.Is(err, syscall.ENOTTY) {
			t.Fatal("Enable to sync logger")
		}
	})

	t.Run("Invalid log level", func(t *testing.T) {
		logger, err := NewLogger("invalid")
		assert.Error(t, err)
		assert.Nil(t, logger)
		assert.Equal(t, "logger parse level failed", err.Error())
	})

	t.Run("Logger build failure", func(t *testing.T) {
	})
}
