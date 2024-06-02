package storages

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/Erlast/short-url.git/internal/app/helpers"
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
short VARCHAR(255) NOT NULL, 
    original TEXT NOT NULL)`)

	if err != nil {
		return nil, fmt.Errorf("failed to create table short_urls: %w", err)
	}

	return &PgStorage{db: db}, nil
}

func (pgs *PgStorage) SaveURL(id string, originalURL string) error {
	_, err := pgs.db.Exec("INSERT INTO short_urls(short, original) VALUES ($1, $2)", id, originalURL)
	if err != nil {
		return fmt.Errorf("unable to save url: %w", err)
	}
	return nil
}

func (pgs *PgStorage) GetByID(id string) (string, error) {
	var originalURL string
	err := pgs.db.QueryRow("SELECT original FROM short_urls WHERE short = $1", id).Scan(&originalURL)
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
	err := pgs.db.QueryRow("SELECT count(original) FROM short_urls WHERE short = $1", key).Scan(&count)
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

func (pgs *PgStorage) Save(incoming []helpers.Incoming, baseURL string) ([]helpers.Output, error) {
	length := len(incoming)

	if length == 0 {
		return nil, errors.New("no incoming URLs found")
	}

	result := make([]helpers.Output, 0, length)

	tx, err := pgs.db.Begin()

	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	stmt, err := tx.Prepare("INSERT INTO short_urls(short, original) VALUES ($1,$2)")

	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer func() {
		err := stmt.Close()
		if err != nil {
			log.Printf("failed to close statment: %v", err)
		}
	}()

	for _, item := range incoming {
		_, err := stmt.Exec(&item.CorrelationID, &item.OriginalURL)
		if err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("failed to insert url: %w", err)
		}

		str, err := url.JoinPath(baseURL, "/", item.CorrelationID)

		if err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("failed to join path: %w", err)
		}

		result = append(result, helpers.Output{ShortURL: str, CorrelationID: item.CorrelationID})
	}
	err = tx.Commit()

	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
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
