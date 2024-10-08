package storages

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"sync"

	"go.uber.org/zap"

	"github.com/Erlast/short-url.git/internal/app/helpers"
)

// MemoryStorage харнилище памяти.
type MemoryStorage struct {
	urls map[string]ShortenURL
}

// NewMemoryStorage инициализация хранилища в памяти.
func NewMemoryStorage(_ context.Context) (*MemoryStorage, error) {
	store := &MemoryStorage{urls: map[string]ShortenURL{}}
	return store, nil
}

// SaveURL сохраняет оригинальный URL
//
// Аргументы
//   - ctx: контектс выполнения
//   - originalURL: оригинальный URL
//
// Возвращает
//   - string: сокращенный URL
//   - error: ошибка выполнения
func (s *MemoryStorage) SaveURL(ctx context.Context, originalURL string) (string, error) {
	var shortURL string
	for range 3 {
		rndString := helpers.RandomString(helpers.LenString)

		if !s.IsExists(ctx, rndString) {
			shortURL = rndString
			continue
		}
		return "", errors.New("failed to generate short url")
	}
	uuid := len(s.urls) + 1
	s.urls[shortURL] = ShortenURL{ctx.Value(helpers.UserID), originalURL, shortURL, uuid, false}

	return shortURL, nil
}

// GetByID получение оригинального URL по короткой ссылке
//
// Аргументы
//   - ctx: контектс выполнения
//   - id: короткая ссылка
//
// Возвращает
//   - string: оригинальный URL
//   - error: ошибка выполнения
func (s *MemoryStorage) GetByID(_ context.Context, id string) (string, error) {
	result, ok := s.urls[id]

	if !ok {
		return "", fmt.Errorf("short URL %s was not found", id)
	}

	if result.IsDeleted {
		return "", &helpers.ConflictError{
			Err: helpers.NewIsDeletedErr("short url is deleted"),
		}
	}

	return result.OriginalURL, nil
}

// LoadURLs сохраняет список оригинальных URL
//
// Аргументы
//   - ctx: контектс выполнения
//   - incoming[]: список оригинальных URL
//   - baseURL: базовый URL приложения
//
// Возвращает
//   - output[]: список сокращенных URL
//   - error: ошибка выполнения
func (s *MemoryStorage) LoadURLs(
	ctx context.Context,
	incoming []Incoming,
	baseURL string,
) ([]Output, error) {
	outputs := make([]Output, 0, len(incoming))

	for _, v := range incoming {
		short, err := s.SaveURL(ctx, v.OriginalURL)
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

// IsExists проверка существования URL
//
// Аргументы
//   - ctx: контектс выполнения
//   - key: сокращенный URL
//
// Возвращает
//   - bool: true - сслыка существует, false - ссылка не существует
func (s *MemoryStorage) IsExists(_ context.Context, key string) bool {
	_, ok := s.urls[key]
	return ok
}

// GetUserURLs получение списка оригинальных URL пользователя
//
// Аргументы
//   - ctx: контектс выполнения
//   - baseURL: базовый URL приложения
//
// Возвращает
//   - userURLs[]: список сокращенных URL
//   - error: ошибка выполнения
func (s *MemoryStorage) GetUserURLs(ctx context.Context, baseURL string) ([]UserURLs, error) {
	var result []UserURLs

	for _, v := range s.urls {
		if v.UserID == ctx.Value(helpers.UserID) && !v.IsDeleted {
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

// DeleteUserURLs удаляет спислок URL по переданному списку
//
// Аргументы
//   - ctx: контектс выполнения
//   - listDeleted[]: список URL на удаление
//   - logger: логгер
//
// Возвращает
//   - error: ошибка выполнения
func (s *MemoryStorage) DeleteUserURLs(
	ctx context.Context,
	listDeleted []string,
	logger *zap.SugaredLogger,
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
			if result.UserID == ctx.Value(helpers.UserID) {
				result.IsDeleted = true
				s.urls[v] = result
			}
		}()
	}
	wg.Wait()

	return nil
}

// DeleteHard удаляет URL которые ранее были мягко удалены
//
// Аргументы
//   - ctx: контектс выполнения
//
// Возвращает
//   - error: ошибка выполнения
func (s *MemoryStorage) DeleteHard(_ context.Context) error {
	var result []ShortenURL

	for _, v := range s.urls {
		if !v.IsDeleted {
			result = append(result, v)
		}
	}

	s.urls = make(map[string]ShortenURL)
	for _, v := range result {
		s.urls[strconv.Itoa(v.ID)] = v
	}

	return nil
}
