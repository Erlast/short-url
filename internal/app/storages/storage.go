package storages

import (
	"context"

	"go.uber.org/zap"

	"github.com/Erlast/short-url.git/internal/app/config"
)

// Output структура ответа при массовом сохранении ссылок
type Output struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// Incoming структура тела запроса при массовом сохранении ссылок
type Incoming struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// ShortenURL структура ссылки
type ShortenURL struct {
	UserID      any    `json:"user_id"`
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
	ID          int    `json:"uuid"`
	IsDeleted   bool   `json:"is_deleted"`
}

// UserURLs структура пользловательской ссылки
type UserURLs struct {
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

// URLStorage интерфейс хранилища
type URLStorage interface {
	SaveURL(ctx context.Context, originalURL string) (string, error)
	GetByID(ctx context.Context, id string) (string, error)
	IsExists(ctx context.Context, key string) bool
	LoadURLs(context.Context, []Incoming, string) ([]Output, error)
	GetUserURLs(ctx context.Context, baseURL string) ([]UserURLs, error)
	DeleteUserURLs(ctx context.Context, listDeleted []string, logger *zap.SugaredLogger) error
	DeleteHard(ctx context.Context) error
}

// NewStorage инициализация хранилища в зависимости от настроек приложения.
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
