package storages

import (
	"github.com/Erlast/short-url.git/internal/app/config"
	"go.uber.org/zap"
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

func NewStorage(cfg *config.Cfg, logger *zap.SugaredLogger) (URLStorage, error) {
	switch {
	case cfg.DatabaseDSN != "":
		return NewPgStorage(cfg.DatabaseDSN)
	case cfg.FileStorage != "":
		return NewFileStorage(cfg.FileStorage, logger)
	default:
		return NewMemoryStorage()
	}
}
