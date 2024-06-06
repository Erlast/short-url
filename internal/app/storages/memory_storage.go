package storages

import (
	"context"
	"fmt"
)

type MemoryStorage struct {
	urls map[string]ShortenURL
}

func NewMemoryStorage(_ context.Context) (*MemoryStorage, error) {
	store := &MemoryStorage{urls: map[string]ShortenURL{}}
	return store, nil
}
func (s *MemoryStorage) SaveURL(_ context.Context, id string, originalURL string) error {
	uuid := len(s.urls) + 1
	s.urls[id] = ShortenURL{originalURL, id, uuid}

	return nil
}

func (s *MemoryStorage) GetByID(_ context.Context, id string) (string, error) {
	result, ok := s.urls[id]

	if !ok {
		return "", fmt.Errorf("short URL %s was not found", id)
	}

	return result.OriginalURL, nil
}

func (s *MemoryStorage) Save(_ context.Context, incoming []Incoming, baseURL string) ([]Output, error) {
	return make([]Output, 0, 1), nil
}

func (s *MemoryStorage) IsExists(_ context.Context, key string) bool {
	_, ok := s.urls[key]
	return ok
}
