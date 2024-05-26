package storages

import (
	"fmt"
	"slices"
)

type MemoryStorage struct {
	urls []ShortenURL
}

func (s *MemoryStorage) SaveURL(id string, originalURL string) error {
	uuid := len(s.urls) + 1
	s.urls = append(s.urls, ShortenURL{originalURL, id, uuid})

	return nil
}

func (s *MemoryStorage) GetByID(id string) (string, error) {
	for i := range s.urls {
		if s.urls[i].ShortURL == id {
			return s.urls[i].OriginalURL, nil
		}
	}

	return "", fmt.Errorf("short URL %s was not found", id)
}

func (s *MemoryStorage) IsExists(key string) bool {
	idx := slices.IndexFunc(s.urls, func(shorten ShortenURL) bool { return shorten.ShortURL == key })
	return idx != -1
}
