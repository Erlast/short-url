package storages

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"go.uber.org/zap"
)

const perm600 = 0o600
const perm777 = 0o777

type FileStorage struct {
	memoryStorage *MemoryStorage
	fileStorage   string
}

func (s *FileStorage) Save(fname string) error {
	data, err := json.MarshalIndent(s.memoryStorage.urls, "", "   ")
	if err != nil {
		return errors.New("marshal indent error")
	}

	err = os.WriteFile(fname, data, perm600)
	if err != nil {
		return errors.New("error write file")
	}
	return nil
}

func (s *FileStorage) Load(fname string, logger *zap.SugaredLogger) error {
	err := createFileIfNotExists(fname, s, logger)

	if err != nil {
		return err
	}

	data, err := os.ReadFile(fname)
	if err != nil {
		return errors.New("unable to read file")
	}

	err = json.Unmarshal(data, &s.memoryStorage.urls)
	if err != nil {
		return errors.New("unable to unmarshal")
	}

	return nil
}

func LoadStorageFromFile(storage *FileStorage, logger *zap.SugaredLogger) (*FileStorage, error) {
	fname := storage.fileStorage

	if err := storage.Load(fname, logger); err != nil {
		return &FileStorage{}, err
	}

	return storage, nil
}

func (s *FileStorage) SaveURL(id string, originalURL string) error {
	err := s.memoryStorage.SaveURL(id, originalURL)
	if err != nil {
		return errors.New("unable to save storage")
	}

	err = SaveToFileStorage(s.fileStorage, s)
	if err != nil {
		return errors.New("unable to save storage")
	}

	return nil
}

func (s *FileStorage) GetByID(id string) (string, error) {
	return s.memoryStorage.GetByID(id)
}

func (s *FileStorage) IsExists(key string) bool {
	return s.memoryStorage.IsExists(key)
}

func SaveToFileStorage(fname string, storage *FileStorage) error {
	if err := storage.Save(fname); err != nil {
		return err
	}

	return nil
}

func createFileIfNotExists(fname string, s *FileStorage, logger *zap.SugaredLogger) error {
	_, err := os.Stat(filepath.Dir(fname))

	if os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(fname), perm777)

		if err != nil {
			logger.Error("err:", err)
			return errors.New("can't create directory")
		}
	}

	_, err = os.Stat(fname)
	if os.IsNotExist(err) {
		file, err := os.OpenFile(fname, os.O_CREATE|os.O_RDONLY, perm777)
		if err != nil {
			logger.Error("err:", err)
			return errors.New("can't create file path")
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				logger.Error("unable to close file")
			}
		}(file)

		if err := s.Save(fname); err != nil {
			return err
		}
	}

	return nil
}
