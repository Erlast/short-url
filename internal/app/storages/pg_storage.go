package storages

import (
	"database/sql"
	"fmt"
)

type PgStorage struct {
	db *sql.DB
}

func NewPgStorage(dsn string) (*PgStorage, error) {
	db, err := sql.Open("pgx", dsn)

	if err != nil {
		return nil, fmt.Errorf("unable to connect database: %w", err)
	}

	_, err = db.Exec(`CREATE TABLE 
    IF NOT EXISTS short_urls 
(id SERIAL PRIMARY KEY, 
short_url VARCHAR(10) NOT NULL, 
    original_url TEXT NOT NULL)`)

	if err != nil {
		return nil, fmt.Errorf("failed to create table short_urls: %w", err)
	}

	return &PgStorage{db: db}, nil
}

func (pgs *PgStorage) SaveURL(id string, originalURL string) error {
	_, err := pgs.db.Exec("INSERT INTO short_urls(short_url, original_url) VALUES ($1, $2)", id, originalURL)
	if err != nil {
		return fmt.Errorf("unable to save url: %w", err)
	}
	return nil
}

func (pgs *PgStorage) GetByID(id string) (string, error) {
	var originalURL string
	err := pgs.db.QueryRow("SELECT original_url FROM short_urls WHERE short_url = $1", id).Scan(&originalURL)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("short URL not found %w", err)
		}
		return "", fmt.Errorf("failed to get query: %w", err)
	}
	return originalURL, nil
}

func (pgs *PgStorage) IsExists(key string) bool {
	var count int
	err := pgs.db.QueryRow("SELECT count(original_url) FROM short_urls WHERE short_url = $1", key).Scan(&count)
	if err != nil {
		_ = fmt.Errorf("failed to get query: %w", err)
	}
	return count != 0
}

func (pgs *PgStorage) CheckPing() error {
	err := pgs.db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping db: %w", err)
	}
	return nil
}

func (pgs *PgStorage) Close() error {
	if pgs.db == nil {
		return nil
	}

	err := pgs.db.Close()
	if err != nil {
		return fmt.Errorf("error closing database connection: %w", err)
	}

	return nil
}
