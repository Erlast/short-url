package storages

import (
	"context"
	"os"
	"testing"

	"github.com/Erlast/short-url.git/internal/app/helpers"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMemoryStorage_SaveURL(t *testing.T) {
	ctx := context.WithValue(context.Background(), helpers.UserID, "user1")
	storage, _ := NewMemoryStorage(ctx)

	originalURL := "https://example.com"
	shortURL, err := storage.SaveURL(ctx, originalURL)

	assert.NoError(t, err)
	assert.NotEmpty(t, shortURL)

	retrievedURL, err := storage.GetByID(ctx, shortURL)
	assert.NoError(t, err)
	assert.Equal(t, originalURL, retrievedURL)
}

func TestMemoryStorage_GetByID(t *testing.T) {
	ctx := context.WithValue(context.Background(), helpers.UserID, "user1")
	storage, _ := NewMemoryStorage(ctx)

	originalURL := "https://example.com"
	shortURL, _ := storage.SaveURL(ctx, originalURL)

	retrievedURL, err := storage.GetByID(ctx, shortURL)
	assert.NoError(t, err)
	assert.Equal(t, originalURL, retrievedURL)

	_, err = storage.GetByID(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestMemoryStorage_LoadURLs(t *testing.T) {
	ctx := context.WithValue(context.Background(), helpers.UserID, "user1")
	storage, _ := NewMemoryStorage(ctx)

	incoming := []Incoming{
		{OriginalURL: "https://example1.com", CorrelationID: "1"},
		{OriginalURL: "https://example2.com", CorrelationID: "2"},
	}
	baseURL := "https://short.ly"

	outputs, err := storage.LoadURLs(ctx, incoming, baseURL)
	assert.NoError(t, err)
	assert.Len(t, outputs, len(incoming))

	for i, output := range outputs {
		assert.Contains(t, output.ShortURL, baseURL)
		assert.Equal(t, incoming[i].CorrelationID, output.CorrelationID)
	}
}

func TestMemoryStorage_IsExists(t *testing.T) {
	ctx := context.WithValue(context.Background(), helpers.UserID, "user1")
	storage, _ := NewMemoryStorage(ctx)

	shortURL, _ := storage.SaveURL(ctx, "https://example.com")

	assert.True(t, storage.IsExists(ctx, shortURL))
	assert.False(t, storage.IsExists(ctx, "nonexistent"))
}

func TestMemoryStorage_GetUserURLs(t *testing.T) {
	ctx := context.WithValue(context.Background(), helpers.UserID, "user1")
	storage, _ := NewMemoryStorage(ctx)

	_, err := storage.SaveURL(ctx, "https://example1.com")
	if err != nil {
		return
	}
	_, err = storage.SaveURL(ctx, "https://example2.com")
	if err != nil {
		return
	}

	userURLs, err := storage.GetUserURLs(ctx, "https://short.ly")
	assert.NoError(t, err)
	assert.Len(t, userURLs, 2)
}

func TestMemoryStorage_DeleteUserURLs(t *testing.T) {
	ctx := context.WithValue(context.Background(), helpers.UserID, "user1")
	storage, _ := NewMemoryStorage(ctx)

	logger, _ := zap.NewDevelopment()

	shortURL1, _ := storage.SaveURL(ctx, "https://example1.com")
	shortURL2, _ := storage.SaveURL(ctx, "https://example2.com")

	err := storage.DeleteUserURLs(ctx, []string{shortURL1, shortURL2}, logger.Sugar())
	assert.NoError(t, err)

	_, err = storage.GetByID(ctx, shortURL1)
	assert.Error(t, err)

	_, err = storage.GetByID(ctx, shortURL2)
	assert.Error(t, err)
}

func TestMemoryStorage_DeleteHard(t *testing.T) {
	ctx := context.WithValue(context.Background(), helpers.UserID, "user1")
	storage, _ := NewMemoryStorage(ctx)

	shortURL1, _ := storage.SaveURL(ctx, "https://example1.com")
	shortURL2, _ := storage.SaveURL(ctx, "https://example2.com")

	err := storage.DeleteUserURLs(ctx, []string{shortURL1, shortURL2}, zap.S().With("test", "some"))
	if err != nil {
		return
	}

	err = storage.DeleteHard(ctx)
	assert.NoError(t, err)

	_, err = storage.GetByID(ctx, shortURL1)
	assert.Error(t, err)

	_, err = storage.GetByID(ctx, shortURL2)
	assert.Error(t, err)
}

func setupTestFileStorage(t *testing.T) (*FileStorage, func()) {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	filePath := "test_storage.json"

	storage, err := NewFileStorage(context.Background(), filePath, logger.Sugar())
	assert.NoError(t, err)

	return storage, func() {
		_ = os.Remove(filePath)
	}
}

func TestFileStorage_SaveURL(t *testing.T) {
	storage, cleanup := setupTestFileStorage(t)
	defer cleanup()

	ctx := context.WithValue(context.Background(), helpers.UserID, "user1")
	originalURL := "https://example.com"
	shortURL, err := storage.SaveURL(ctx, originalURL)

	assert.NoError(t, err)
	assert.NotEmpty(t, shortURL)

	retrievedURL, err := storage.GetByID(ctx, shortURL)
	assert.NoError(t, err)
	assert.Equal(t, originalURL, retrievedURL)
}

func TestFileStorage_LoadURLs(t *testing.T) {
	storage, cleanup := setupTestFileStorage(t)
	defer cleanup()

	ctx := context.WithValue(context.Background(), helpers.UserID, "user1")
	incoming := []Incoming{
		{OriginalURL: "https://example1.com", CorrelationID: "1"},
		{OriginalURL: "https://example2.com", CorrelationID: "2"},
	}
	baseURL := "https://short.ly"

	outputs, err := storage.LoadURLs(ctx, incoming, baseURL)
	assert.NoError(t, err)
	assert.Len(t, outputs, len(incoming))

	for i, output := range outputs {
		assert.Contains(t, output.ShortURL, baseURL)
		assert.Equal(t, incoming[i].CorrelationID, output.CorrelationID)
	}
}

func TestFileStorage_DeleteUserURLs(t *testing.T) {
	storage, cleanup := setupTestFileStorage(t)
	defer cleanup()

	ctx := context.WithValue(context.Background(), helpers.UserID, "user1")
	logger, _ := zap.NewDevelopment()

	shortURL1, _ := storage.SaveURL(ctx, "https://example1.com")
	shortURL2, _ := storage.SaveURL(ctx, "https://example2.com")

	err := storage.DeleteUserURLs(ctx, []string{shortURL1, shortURL2}, logger.Sugar())
	assert.NoError(t, err)

	_, err = storage.GetByID(ctx, shortURL1)
	assert.Error(t, err)

	_, err = storage.GetByID(ctx, shortURL2)
	assert.Error(t, err)
}

func TestFileStorage_DeleteHard(t *testing.T) {
	storage, cleanup := setupTestFileStorage(t)
	defer cleanup()

	ctx := context.WithValue(context.Background(), helpers.UserID, "user1")

	shortURL1, _ := storage.SaveURL(ctx, "https://example1.com")
	shortURL2, _ := storage.SaveURL(ctx, "https://example2.com")

	err := storage.DeleteUserURLs(ctx, []string{shortURL1, shortURL2}, zap.S().With("test", "somearg"))
	if err != nil {
		return
	}

	err = storage.DeleteHard(ctx)
	assert.NoError(t, err)

	_, err = storage.GetByID(ctx, shortURL1)
	assert.Error(t, err)

	_, err = storage.GetByID(ctx, shortURL2)
	assert.Error(t, err)
}

func TestFileStorage_Persistence(t *testing.T) {
	filePath := "test_storage.json"
	logger, _ := zap.NewDevelopment()

	storage, err := NewFileStorage(context.Background(), filePath, logger.Sugar())
	assert.NoError(t, err)

	ctx := context.WithValue(context.Background(), helpers.UserID, "user1")
	originalURL := "https://example.com"
	shortURL, err := storage.SaveURL(ctx, originalURL)
	assert.NoError(t, err)

	// Re-load the storage
	storage, err = NewFileStorage(context.Background(), filePath, logger.Sugar())
	assert.NoError(t, err)

	// Verify the URL is still there after reload
	retrievedURL, err := storage.GetByID(ctx, shortURL)
	assert.NoError(t, err)
	assert.Equal(t, originalURL, retrievedURL)

	_ = os.Remove(filePath)
}
