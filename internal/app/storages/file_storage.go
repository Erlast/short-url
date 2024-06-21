package storages

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

const perm600 = 0o600
const perm777 = 0o777

type FileStorage struct {
	*MemoryStorage
	logger      *zap.SugaredLogger
	fileStorage string
}

func NewFileStorage(_ context.Context, fileStorage string, logger *zap.SugaredLogger) (*FileStorage, error) {
	storage, err := loadStorageFromFile(
		&FileStorage{
			&MemoryStorage{
				urls: map[string]ShortenURL{},
			},
			logger,
			fileStorage},
		logger)
	if err != nil {
		return nil, errors.New("unable to load storage")
	}
	return storage, nil
}

func (s *FileStorage) SaveURL(ctx context.Context, id string, originalURL string, user *CurrentUser) error {
	err := s.MemoryStorage.SaveURL(ctx, id, originalURL, user)
	if err != nil {
		return errors.New("unable to save storage")
	}

	urls := make([]ShortenURL, 0, len(s.MemoryStorage.urls))
	for _, value := range s.MemoryStorage.urls {
		urls = append(urls, ShortenURL{user, value.OriginalURL, value.ShortURL, value.ID, false})
	}
	err = saveToFileStorage(s, &urls)
	if err != nil {
		return errors.New("unable to save storage")
	}

	return nil
}

func (s *FileStorage) LoadURLs(
	ctx context.Context,
	incoming []Incoming,
	baseURL string,
	user *CurrentUser,
) ([]Output, error) {
	err := s.load(s.fileStorage, s.logger)
	if err != nil {
		return nil, errors.New("unable to load storage")
	}

	outputs, err := s.MemoryStorage.LoadURLs(ctx, incoming, baseURL, user)

	if err != nil {
		return nil, fmt.Errorf("error saving batch infile: %w", err)
	}
	var urls = make([]ShortenURL, 0, len(s.MemoryStorage.urls))
	for _, value := range s.MemoryStorage.urls {
		lenItems := len(urls) + 1
		urls = append(urls, ShortenURL{user, value.OriginalURL, value.ShortURL, lenItems, value.IsDeleted})
	}

	err = s.save(&urls)
	if err != nil {
		return nil, fmt.Errorf("error saving batch infile: %w", err)
	}

	return outputs, nil
}

func (s *FileStorage) DeleteUserURLs(
	ctx context.Context,
	listDeleted []string,
	logger *zap.SugaredLogger,
	user *CurrentUser,
) error {
	err := s.MemoryStorage.DeleteUserURLs(ctx, listDeleted, logger, user)
	if err != nil {
		return errors.New("unable to delete users")
	}
	var urls = make([]ShortenURL, 0, len(s.MemoryStorage.urls))
	for _, value := range s.MemoryStorage.urls {
		lenItems := len(urls) + 1
		urls = append(urls, ShortenURL{user, value.OriginalURL, value.ShortURL, lenItems, value.IsDeleted})
	}
	err = s.save(&urls)
	if err != nil {
		return fmt.Errorf("error saving batch infile: %w", err)
	}
	return nil
}

func saveToFileStorage(s *FileStorage, url *[]ShortenURL) error {
	if err := s.save(url); err != nil {
		return err
	}

	return nil
}

func loadStorageFromFile(storage *FileStorage, logger *zap.SugaredLogger) (*FileStorage, error) {
	fname := storage.fileStorage

	if err := storage.load(fname, logger); err != nil {
		return &FileStorage{}, err
	}

	return storage, nil
}

func (s *FileStorage) save(urls *[]ShortenURL) error {
	data, err := json.MarshalIndent(urls, "", "   ")
	if err != nil {
		return errors.New("marshal indent error")
	}

	err = os.WriteFile(s.fileStorage, data, perm600)
	if err != nil {
		return fmt.Errorf("unable to read file: %w", err)
	}
	return nil
}

func (s *FileStorage) load(fname string, logger *zap.SugaredLogger) error {
	err := createFileIfNotExists(fname, s, logger)

	if err != nil {
		return err
	}

	data, err := os.ReadFile(fname)
	if err != nil {
		return fmt.Errorf("unable to read file: %w", err)
	}
	var urls []ShortenURL
	err = json.Unmarshal(data, &urls)
	if err != nil {
		return errors.New("unable to unmarshal")
	}

	for _, v := range urls {
		s.MemoryStorage.urls[v.ShortURL] = v
	}

	return nil
}

func createFileIfNotExists(fname string, s *FileStorage, logger *zap.SugaredLogger) error {
	_, err := os.Stat(filepath.Dir(fname))

	if os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(fname), perm777)

		if err != nil {
			return errors.New("can't create directory")
		}
	}

	_, err = os.Stat(fname)
	if os.IsNotExist(err) {
		file, err := os.OpenFile(fname, os.O_CREATE|os.O_RDONLY, perm777)
		if err != nil {
			return errors.New("can't create file path")
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				logger.Error("unable to close file: ", err)
			}
		}(file)

		var urls []ShortenURL
		if err := s.save(&urls); err != nil {
			return err
		}
	}

	return nil
}
