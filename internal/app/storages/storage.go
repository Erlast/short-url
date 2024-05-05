package storages

import "errors"

type Storage struct {
	urls map[string]string
}

func Init() *Storage {
	return &Storage{urls: make(map[string]string)}
}

func (s *Storage) SaveURL(id string, originalURL string) {
	s.urls[id] = originalURL
}

func (s *Storage) GetByID(id string) (string, error) {
	originalURL, ok := s.urls[id]

	if !ok {
		return "", errors.New("not found")
	}

	return originalURL, nil
}
