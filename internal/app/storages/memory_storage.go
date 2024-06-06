package storages

import (
	"context"
	"fmt"
	"net/url"
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

func (s *MemoryStorage) LoadURLs(ctx context.Context, incoming []Incoming, baseURL string) ([]Output, error) {
	outputs := make([]Output, 0, len(incoming))

	for _, v := range incoming {
		err := s.SaveURL(ctx, v.CorrelationID, v.OriginalURL)
		if err != nil {
			return nil, fmt.Errorf("save batch error: %w", err)
		}

		shortURL, err := url.JoinPath(baseURL, "/", v.CorrelationID)

		if err != nil {
			return nil, fmt.Errorf("error getFullShortURL from two parts %w", err)
		}

		outputs = append(outputs, Output{
			CorrelationID: v.CorrelationID,
			ShortURL:      shortURL,
		})
	}

	return outputs, nil
}

func (s *MemoryStorage) IsExists(_ context.Context, key string) bool {
	_, ok := s.urls[key]
	return ok
}
