package storages

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const perm600 = 0o600
const perm777 = 0o777

type FileStorage struct {
	memoryStorage *MemoryStorage
	fileStorage   string
}

func NewFileStorage(fileStorage string) (*FileStorage, error) {
	storage, err := loadStorageFromFile(&FileStorage{memoryStorage: &MemoryStorage{urls: []ShortenURL{}}, fileStorage: fileStorage})
	if err != nil {
		return nil, errors.New("unable to load storage")
	}
	return storage, nil
}

func (s *FileStorage) SaveURL(id string, originalURL string) error {
	err := s.memoryStorage.SaveURL(id, originalURL)
	if err != nil {
		return errors.New("unable to save storage")
	}

	err = saveToFileStorage(s.fileStorage, s)
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

func saveToFileStorage(fname string, storage *FileStorage) error {
	if err := storage.save(fname); err != nil {
		return err
	}

	return nil
}

func loadStorageFromFile(storage *FileStorage) (*FileStorage, error) {
	fname := storage.fileStorage

	if err := storage.load(fname); err != nil {
		return &FileStorage{}, err
	}

	return storage, nil
}

func (s *FileStorage) save(fname string) error {
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

func (s *FileStorage) load(fname string) error {
	err := createFileIfNotExists(fname, s)

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

func createFileIfNotExists(fname string, s *FileStorage) error {
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
				fmt.Println("unable to close file")
			}
		}(file)

		if err := s.save(fname); err != nil {
			return err
		}
	}

	return nil
}
