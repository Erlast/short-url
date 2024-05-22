package storages

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/Erlast/short-url.git/internal/app/config"
	"github.com/Erlast/short-url.git/internal/app/logger"
)

const perm600 = 0o600
const perm777 = 0o777

func (s *Storage) Save(fname string) error {
	path := getFullFilePath(fname)

	data, err := json.MarshalIndent(s.urls, "", "   ")
	if err != nil {
		return errors.New("marshal indent error")
	}

	err = os.WriteFile(path, data, perm600)
	if err != nil {
		return errors.New("error write file")
	}
	return nil
}

func getFullFilePath(fname string) string {
	tempPath := os.TempDir()

	path := filepath.Join(tempPath, fname)
	return path
}

func (s *Storage) Load(fname string) error {
	path := getFullFilePath(fname)

	_, err := os.Stat(filepath.Dir(path))

	if os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(path), perm777)
		if err != nil {
			logger.Log.Error("can't create directory", err)
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, perm777)
		if err != nil {
			logger.Log.Error("can't create file path", err)
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				logger.Log.Error("unable to close file")
			}
		}(file)
	}

	if err := s.Save(fname); err != nil {
		return err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return errors.New("unable to read file")
	}

	err = json.Unmarshal(data, &s.urls)
	if err != nil {
		return errors.New("unable to unmarshal")
	}

	return nil
}

func LoadStorageFromFile(cfg *config.Cfg, storage *Storage) (*Storage, error) {
	fname := cfg.FileStorage

	if err := storage.Load(fname); err != nil {
		return &Storage{}, err
	}

	return storage, nil
}

func SaveToFileStorage(fname string, storage *Storage) error {
	if err := storage.Save(fname); err != nil {
		return err
	}

	return nil
}
