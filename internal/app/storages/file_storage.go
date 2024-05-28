package storages

import (
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
	fileStorage string
}

func NewFileStorage(fileStorage string, logger *zap.SugaredLogger) (*FileStorage, error) {
	storage, err := loadStorageFromFile(&FileStorage{fileStorage: fileStorage}, logger)
	if err != nil {
		return nil, errors.New("unable to load storage")
	}
	return storage, nil
}

func (s *FileStorage) SaveURL(id string, originalURL string) error {
	err := s.MemoryStorage.SaveURL(id, originalURL)
	if err != nil {
		return errors.New("unable to save storage")
	}

	err = saveToFileStorage(s.fileStorage, s)
	if err != nil {
		return errors.New("unable to save storage")
	}

	return nil
}

func saveToFileStorage(fname string, storage *FileStorage) error {
	if err := storage.save(fname); err != nil {
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

func (s *FileStorage) save(fname string) error {
	data, err := json.MarshalIndent(s.urls, "", "   ")
	if err != nil {
		return errors.New("marshal indent error")
	}

	err = os.WriteFile(fname, data, perm600)
	if err != nil {
		return errors.New("error write file")
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

	err = json.Unmarshal(data, &s.urls)
	if err != nil {
		return errors.New("unable to unmarshal")
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

		if err := s.save(fname); err != nil {
			return err
		}
	}

	return nil
}
