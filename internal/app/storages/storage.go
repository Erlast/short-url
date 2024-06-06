package storages

import (
	"context"

	"go.uber.org/zap"

	"github.com/Erlast/short-url.git/internal/app/config"
)

type URLStorage interface {
	SaveURL(ctx context.Context, id string, originalURL string) error
	GetByID(ctx context.Context, id string) (string, error)
	IsExists(ctx context.Context, key string) bool
	Save(context.Context, []Incoming, string) ([]Output, error)
}

type Output struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type Incoming struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type ShortenURL struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	ID          int    `json:"uuid"`
}

func NewStorage(ctx context.Context, cfg *config.Cfg, logger *zap.SugaredLogger) (URLStorage, error) {
	switch {
	case cfg.DatabaseDSN != "":
		return NewPgStorage(ctx, cfg.DatabaseDSN)
	case cfg.FileStorage != "":
		return NewFileStorage(ctx, cfg.FileStorage, logger)
	default:
		return NewMemoryStorage(ctx)
	}
}
