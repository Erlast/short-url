package storages

import (
	"fmt"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/logger"
)

type ShortenURL struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	ID          int    `json:"uuid"`
}

type Storage struct {
	fileStorage string
	urls        []ShortenURL
}

func NewStorage(cfg *config.Cfg) *Storage {
	store := &Storage{urls: []ShortenURL{}, fileStorage: cfg.FileStorage}
	if cfg.FileStorage != "" {
		storage, err := LoadStorageFromFile(store)
		if err != nil {
			logger.Log.Fatal("unable to load storage:", err)
		}
		return storage
	}
	return store
}

func (s *Storage) SaveURL(id string, originalURL string) {
	uuid := len(s.urls) + 1
	s.urls = append(s.urls, ShortenURL{originalURL, id, uuid})
	if s.fileStorage != "" {
		err := SaveToFileStorage(s.fileStorage, s)
		if err != nil {
			logger.Log.Fatal("Unable to save storage:", err)
		}
	}
}

func (s *Storage) GetByID(id string) (string, error) {
	for i := range s.urls {
		if s.urls[i].ShortURL == id {
			return s.urls[i].OriginalURL, nil
		}
	}

	return "", fmt.Errorf("short URL %s was not found", id)
}

func (s *Storage) IsExists(key string) bool {
	for i := range s.urls {
		if s.urls[i].ShortURL == key {
			return true
		}
	}

	return false
}
