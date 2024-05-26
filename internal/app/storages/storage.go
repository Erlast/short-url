package storages

import (
	"go.uber.org/zap"

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

func NewStorage(cfg *config.Cfg, logger *zap.SugaredLogger) URLStorage {
	store := &MemoryStorage{urls: []ShortenURL{}}
	if cfg.FileStorage != "" {
		storage, err := LoadStorageFromFile(&FileStorage{memoryStorage: store, fileStorage: cfg.FileStorage}, logger)
		if err != nil {
			logger.Fatal("unable to load storage: ", err)
		}
		return storage
	}
	return store
}
