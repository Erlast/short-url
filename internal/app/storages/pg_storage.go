package storages

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"net/url"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/Erlast/short-url.git/internal/app/helpers"
)

type PgStorage struct {
	db *pgxpool.Pool
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func NewPgStorage(ctx context.Context, dsn string) (*PgStorage, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}
	conn, err := initPool(ctx, dsn)

	if err != nil {
		return nil, fmt.Errorf("unable to connect database: %w", err)
	}

	go deleteSoftDeletedRecords(ctx, conn)

	return &PgStorage{db: conn}, nil
}

func (pgs *PgStorage) SaveURL(ctx context.Context, id string, originalURL string) error {
	sqlString := "INSERT INTO short_urls(short, original, user_id, is_deleted) VALUES ($1, $2, $3, $4)"
	_, err := pgs.db.Exec(ctx, sqlString, id, originalURL, ctx.Value(helpers.UserID), false)

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
	var isDeleted bool
	err := pgs.db.QueryRow(ctx, "SELECT original, is_deleted FROM short_urls WHERE short = $1", id).Scan(
		&originalURL,
		&isDeleted,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("short URL not found %w", err)
		}
		return "", fmt.Errorf("failed to get query: %w", err)
	}
	if isDeleted {
		return "", &helpers.ConflictError{
			Err: helpers.NewIsDeletedErr("short url is deleted"),
		}
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

func (pgs *PgStorage) LoadURLs(
	ctx context.Context,
	incoming []Incoming,
	baseURL string,
) ([]Output, error) {
	length := len(incoming)

	if length == 0 {
		return nil, errors.New("no incoming URLs found")
	}

	result := make([]Output, 0, length)

	batch := &pgx.Batch{}
	stmt := "INSERT INTO short_urls(short, original, user_id) VALUES (@short,@original,@user_id) returning (short)"

	for _, item := range incoming {
		var shortURL string
		for range 3 {
			rndString := helpers.RandomString(helpers.LenString)

			if !pgs.IsExists(ctx, rndString) {
				shortURL = rndString
				continue
			}
			return nil, errors.New("failed to generate short url")
		}

		args := pgx.NamedArgs{"short": shortURL, "original": item.OriginalURL, "user_id": ctx.Value(helpers.UserID)}
		batch.Queue(stmt, args)
	}

	tx, err := pgs.db.Begin(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if e := tx.Rollback(ctx); e != nil {
			err = fmt.Errorf("failed to rollback the transaction: %w", e)
			return
		}
	}()

	results := tx.SendBatch(ctx, batch)

	defer func() {
		if e := results.Close(); e != nil {
			err = fmt.Errorf("closing batch results error: %w", e)
			return
		}

		if e := tx.Commit(ctx); e != nil {
			err = fmt.Errorf("unable to commit: %w", e)
			return
		}
	}()

	for _, item := range incoming {
		var short string

		err = results.QueryRow().Scan(&short)

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

		str, err := url.JoinPath(baseURL, "/", short)
		if err != nil {
			return nil, fmt.Errorf("unable to create path: %w", err)
		}
		result = append(result, Output{ShortURL: str, CorrelationID: item.CorrelationID})
	}

	return result, nil
}

func (pgs *PgStorage) GetUserURLs(ctx context.Context, baseURL string) ([]UserURLs, error) {
	var result []UserURLs

	sqlSring := "SELECT short, original FROM short_urls WHERE user_id = $1 and is_deleted=false"
	rows, err := pgs.db.Query(ctx, sqlSring, ctx.Value(helpers.UserID))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user URLs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var userURL UserURLs
		if err = rows.Scan(
			&userURL.ShortURL,
			&userURL.OriginalURL,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		shortURL, err := url.JoinPath(baseURL, "/", userURL.ShortURL)
		if err != nil {
			return nil, fmt.Errorf("unable to create path: %w", err)
		}
		userURL.ShortURL = shortURL
		result = append(result, userURL)
	}
	return result, nil
}

func (pgs *PgStorage) DeleteUserURLs(
	ctx context.Context,
	listDeleted []string,
	_ *zap.SugaredLogger,
) error {
	batch := &pgx.Batch{}
	for _, shortURL := range listDeleted {
		batch.Queue(
			"UPDATE short_urls set is_deleted=true WHERE short = $1 and user_id=$2",
			shortURL,
			ctx.Value(helpers.UserID),
		)
	}

	tx, err := pgs.db.Begin(ctx)

	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if e := tx.Rollback(ctx); e != nil {
			err = fmt.Errorf("failed to rollback the transaction: %w", e)
			return
		}
	}()

	results := tx.SendBatch(ctx, batch)

	defer func() {
		if e := results.Close(); e != nil {
			err = fmt.Errorf("closing batch results error: %w", e)
			return
		}

		if e := tx.Commit(ctx); e != nil {
			err = fmt.Errorf("unable to commit: %w", e)
			return
		}
	}()
	return nil
}

func (pgs *PgStorage) Close() error {
	if pgs.db == nil {
		return nil
	}

	pgs.db.Close()

	return nil
}

func initPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the DSN: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize a connection pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping the DB: %w", err)
	}
	return pool, nil
}

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err := m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}
