package storages

import (
	"fmt"
)

type MemoryStorage struct {
	urls map[string]ShortenURL
}

func NewMemoryStorage() (*MemoryStorage, error) {
	store := &MemoryStorage{urls: map[string]ShortenURL{}}
	return store, nil
}
func (s *MemoryStorage) SaveURL(id string, originalURL string) error {
	uuid := len(s.urls) + 1
	s.urls[id] = ShortenURL{originalURL, id, uuid}

	return nil
}

func (s *MemoryStorage) GetByID(id string) (string, error) {
	result, ok := s.urls[id]

	if !ok {
		return "", fmt.Errorf("short URL %s was not found", id)
	}

	return result.OriginalURL, nil
}

func (s *MemoryStorage) IsExists(key string) bool {
	_, ok := s.urls[key]
	return ok
}
