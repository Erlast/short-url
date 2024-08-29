package storages

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMockURLStorage(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockStorage := NewMockURLStorage(ctrl)
	logger := zap.NewNop().Sugar()

	t.Run("SaveURL", func(t *testing.T) {
		mockStorage.EXPECT().
			SaveURL(context.Background(), "http://example.com").
			Return("shortURL", nil).
			Times(1)

		shortURL, err := mockStorage.SaveURL(context.Background(), "http://example.com")
		assert.NoError(t, err)
		assert.Equal(t, "shortURL", shortURL)
	})

	t.Run("GetByID", func(t *testing.T) {
		mockStorage.EXPECT().
			GetByID(context.Background(), "uuid").
			Return("http://example.com", nil).
			Times(1)

		originalURL, err := mockStorage.GetByID(context.Background(), "uuid")
		assert.NoError(t, err)
		assert.Equal(t, "http://example.com", originalURL)
	})

	t.Run("IsExists", func(t *testing.T) {
		mockStorage.EXPECT().
			IsExists(context.Background(), "shortURL").
			Return(true).
			Times(1)

		exists := mockStorage.IsExists(context.Background(), "shortURL")
		assert.True(t, exists)
	})

	t.Run("LoadURLs", func(t *testing.T) {
		incoming := []Incoming{
			{CorrelationID: "123", OriginalURL: "http://example.com"},
		}
		expectedOutput := []Output{
			{CorrelationID: "123", ShortURL: "shortURL"},
		}

		mockStorage.EXPECT().
			LoadURLs(context.Background(), incoming, "baseURL").
			Return(expectedOutput, nil).
			Times(1)

		output, err := mockStorage.LoadURLs(context.Background(), incoming, "baseURL")
		assert.NoError(t, err)
		assert.Equal(t, expectedOutput, output)
	})

	t.Run("GetUserURLs", func(t *testing.T) {
		expectedUserURLs := []UserURLs{
			{OriginalURL: "http://example.com", ShortURL: "shortURL"},
		}

		mockStorage.EXPECT().
			GetUserURLs(context.Background(), "baseURL").
			Return(expectedUserURLs, nil).
			Times(1)

		userURLs, err := mockStorage.GetUserURLs(context.Background(), "baseURL")
		assert.NoError(t, err)
		assert.Equal(t, expectedUserURLs, userURLs)
	})

	t.Run("DeleteUserURLs", func(t *testing.T) {
		listDeleted := []string{"shortURL1", "shortURL2"}

		mockStorage.EXPECT().
			DeleteUserURLs(context.Background(), listDeleted, logger).
			Return(nil).
			Times(1)

		err := mockStorage.DeleteUserURLs(context.Background(), listDeleted, logger)
		assert.NoError(t, err)
	})

	t.Run("DeleteHard", func(t *testing.T) {
		mockStorage.EXPECT().
			DeleteHard(context.Background()).
			Return(nil).
			Times(1)

		err := mockStorage.DeleteHard(context.Background())
		assert.NoError(t, err)
	})
}
