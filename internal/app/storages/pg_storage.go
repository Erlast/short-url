package storages

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Erlast/short-url.git/internal/app/helpers"
)

type PgStorage struct {
	db *pgx.Conn
}

func NewPgStorage(ctx context.Context, dsn string) (*PgStorage, error) {
	conn, err := pgx.Connect(ctx, dsn)

	if err != nil {
		return nil, fmt.Errorf("unable to connect database: %w", err)
	}

	sqlCreate := `CREATE TABLE 
    IF NOT EXISTS short_urls 
(id SERIAL PRIMARY KEY, 
short VARCHAR(255) NOT NULL, 
    original TEXT NOT NULL,
    UNIQUE (original)
    )`
	_, err = conn.Exec(context.Background(), sqlCreate)

	if err != nil {
		return nil, fmt.Errorf("failed to create table short_urls: %w", err)
	}

	return &PgStorage{db: conn}, nil
}

func (pgs *PgStorage) SaveURL(ctx context.Context, id string, originalURL string) error {
	_, err := pgs.db.Exec(ctx, "INSERT INTO short_urls(short, original) VALUES ($1, $2)", id, originalURL)

	if err != nil {
		var pgsErr *pgconn.PgError
		if errors.As(err, &pgsErr) && pgsErr.Code == pgerrcode.UniqueViolation {
			var existingShortURL string
			err = pgs.db.QueryRow(ctx, `
                SELECT short FROM short_urls WHERE original = $1
            `, originalURL).Scan(&existingShortURL)

			if err != nil {
				return fmt.Errorf("falied to get short url: %w", err)
			}
			return &helpers.ConflictError{
				ShortURL: existingShortURL,
				Err:      err,
			}
		}
		return fmt.Errorf("unable to save url: %w", err)
	}
	return nil
}

func (pgs *PgStorage) GetByID(ctx context.Context, id string) (string, error) {
	var originalURL string
	err := pgs.db.QueryRow(ctx, "SELECT original FROM short_urls WHERE short = $1", id).Scan(&originalURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("short URL not found %w", err)
		}
		return "", fmt.Errorf("failed to get query: %w", err)
	}
	return originalURL, nil
}

func (pgs *PgStorage) IsExists(ctx context.Context, key string) bool {
	var count int
	err := pgs.db.QueryRow(ctx, "SELECT count(original) FROM short_urls WHERE short = $1", key).Scan(&count)
	if err != nil {
		_ = fmt.Errorf("failed to get query: %w", err)
	}
	return count != 0
}

func (pgs *PgStorage) CheckPing(ctx context.Context) error {
	err := pgs.db.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping db: %w", err)
	}
	return nil
}

func (pgs *PgStorage) LoadURLs(ctx context.Context, incoming []Incoming, baseURL string) ([]Output, error) {
	length := len(incoming)

	if length == 0 {
		return nil, errors.New("no incoming URLs found")
	}

	result := make([]Output, 0, length)

	tx, err := pgs.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	batch := &pgx.Batch{}

	stmt := "INSERT INTO short_urls(short, original) VALUES (@short,@original)"

	for _, item := range incoming {
		args := pgx.NamedArgs{"short": item.CorrelationID, "original": item.OriginalURL}
		batch.Queue(stmt, args)
	}

	results := tx.SendBatch(ctx, batch)
	defer func() {
		if e := results.Close(); e != nil {
			err = fmt.Errorf("closing batch results: %w", err)
		}

		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			if e := tx.Commit(ctx); e != nil {
				err = fmt.Errorf("unable to commit: %w", err)
			}
		}
	}()

	for _, item := range incoming {
		_, err := results.Exec()
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				return nil, &helpers.ConflictError{
					ShortURL: item.OriginalURL,
					Err:      err,
				}
			}

			return nil, fmt.Errorf("unable to insert row: %w", err)
		}

		str, err := url.JoinPath(baseURL, "/", item.CorrelationID)
		if err != nil {
			return nil, fmt.Errorf("unable to create path: %w", err)
		}
		result = append(result, Output{ShortURL: str, CorrelationID: item.CorrelationID})
	}

	return result, nil
}

func (pgs *PgStorage) Close(ctx context.Context) error {
	if pgs.db == nil {
		return nil
	}

	err := pgs.db.Close(ctx)
	if err != nil {
		return fmt.Errorf("error closing database connection: %w", err)
	}

	return nil
}
