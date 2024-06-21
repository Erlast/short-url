package storages

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sync"

	"go.uber.org/zap"

	"github.com/Erlast/short-url.git/internal/app/helpers"
)

type MemoryStorage struct {
	urls map[string]ShortenURL
}

func NewMemoryStorage(_ context.Context) (*MemoryStorage, error) {
	store := &MemoryStorage{urls: map[string]ShortenURL{}}
	return store, nil
}
func (s *MemoryStorage) SaveURL(_ context.Context, id string, originalURL string, user *CurrentUser) error {
	uuid := len(s.urls) + 1
	s.urls[id] = ShortenURL{user, originalURL, id, uuid, false}

	return nil
}

func (s *MemoryStorage) GetByID(_ context.Context, id string) (string, error) {
	result, ok := s.urls[id]

	if !ok {
		return "", fmt.Errorf("short URL %s was not found", id)
	}

	if result.IsDeleted {
		return "", &helpers.IsDeletedError{
			Err: errors.New("short URL is deleted"),
		}
	}

	return result.OriginalURL, nil
}

func (s *MemoryStorage) LoadURLs(
	ctx context.Context,
	incoming []Incoming,
	baseURL string,
	user *CurrentUser,
) ([]Output, error) {
	outputs := make([]Output, 0, len(incoming))

	for _, v := range incoming {
		var short string
		for range 3 {
			rndString := helpers.RandomString(helpers.LenString)

			if !s.IsExists(ctx, rndString) {
				short = rndString
				continue
			}
			return nil, errors.New("failed to generate short url")
		}
		err := s.SaveURL(ctx, short, v.OriginalURL, user)
		if err != nil {
			return nil, fmt.Errorf("save batch error: %w", err)
		}

		shortURL, err := url.JoinPath(baseURL, "/", short)

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

func (s *MemoryStorage) GetUserURLs(_ context.Context, baseURL string, user *CurrentUser) ([]UserURLs, error) {
	var result []UserURLs

	for _, v := range s.urls {
		if v.User.UserID == user.UserID && !v.IsDeleted {
			shortURL, err := url.JoinPath(baseURL, "/", v.ShortURL)
			if err != nil {
				return nil, fmt.Errorf("error getFullShortURL from two parts %w", err)
			}
			result = append(result, UserURLs{ShortURL: shortURL, OriginalURL: v.OriginalURL})
		}
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}

func (s *MemoryStorage) DeleteUserURLs(
	_ context.Context,
	listDeleted []string,
	logger *zap.SugaredLogger,
	user *CurrentUser,
) error {
	var wg sync.WaitGroup
	for _, v := range listDeleted {
		v := v
		wg.Add(1)
		go func() {
			defer wg.Done()
			result, ok := s.urls[v]
			if !ok {
				logger.Errorf("short URL %s was not found", v)
				return
			}
			if result.User.UserID == user.UserID {
				result.IsDeleted = true
				s.urls[v] = result
			}
		}()
	}
	wg.Wait()

	return nil
}
