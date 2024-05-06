package storages

import (
	"fmt"
)

type Storage struct {
	urls map[string]string
}

func NewStorage() *Storage {
	return &Storage{urls: make(map[string]string)}
}

func (s *Storage) SaveURL(id string, originalURL string) {
	s.urls[id] = originalURL
}

func (s *Storage) GetByID(id string) (string, error) {
	originalURL, ok := s.urls[id]

	if !ok {
		return "", fmt.Errorf("short URL %s was not found", id)
	}

	return originalURL, nil
}
