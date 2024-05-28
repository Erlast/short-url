package storages

import (
	"github.com/Erlast/short-url.git/internal/app/config"
)

type URLStorage interface {
	SaveURL(id string, originalURL string) error
	GetByID(id string) (string, error)
	IsExists(key string) bool
}

type ShortenURL struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	ID          int    `json:"uuid"`
}

func NewStorage(cfg *config.Cfg) (URLStorage, error) {
	if cfg.FileStorage != "" {
		return NewFileStorage(cfg.FileStorage)
	}
	return NewMemoryStorage()
}
