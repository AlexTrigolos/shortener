package model

import (
	"context"
	"database/sql"
	"errors"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBURL struct {
	db *sql.DB
}
type MigrationService interface {
	RunMigrations(db *sql.DB) (*database.Driver, *migrate.Migrate, error)
}

func NewDBURL(db *sql.DB, migration MigrationService) (*DBURL, *database.Driver, *migrate.Migrate, error) {
	dbDrv, m, err := migration.RunMigrations(db)
	return &DBURL{db: db}, dbDrv, m, err
}

func (dbu *DBURL) Set(ctx context.Context, key, value, userID string) error {
	_, err := dbu.db.ExecContext(ctx, "INSERT INTO urls (short, original, user_id) VALUES ($1, $2, $3)", key, value, userID)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return err
	}
	if pgErr.Code != pgerrcode.UniqueViolation {
		return err
	}

	if pgErr.ConstraintName == "idx_urls_original" {
		return ErrConflict
	}
	return err
}

func (dbu *DBURL) Get(ctx context.Context, key string) (string, error) {
	row := dbu.db.QueryRowContext(ctx, "SELECT original FROM urls WHERE short = $1", key)
	var original string

	err := row.Scan(&original)
	if err != nil {
		return "", err
	}
	return original, nil
}

func (dbu *DBURL) GetByOriginal(ctx context.Context, original, baseURL string) (string, error) {
	row := dbu.db.QueryRowContext(ctx, "SELECT short FROM urls WHERE original = $1", original)
	var short string

	err := row.Scan(&short)
	if err != nil {
		return "", err
	}
	return baseURL + "/" + short, nil
}

func (dbu *DBURL) GetByUserID(ctx context.Context, userID, baseURL string) (*[]UserURL, error) {
	rows, err := dbu.db.QueryContext(ctx, "SELECT short, original FROM urls WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []UserURL
	for rows.Next() {
		var url UserURL
		if err := rows.Scan(&url.ShortURL, &url.OriginalURL); err != nil {
			return nil, err
		}
		url.ShortURL = baseURL + "/" + url.ShortURL
		urls = append(urls, url)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return &urls, nil
}

func (dbu *DBURL) PingContext(ctx context.Context) error {
	return dbu.db.PingContext(ctx)
}

func (dbu *DBURL) Begin(_ int) (*Transaction, error) {
	tx, err := dbu.db.Begin()
	if err != nil {
		return nil, err
	}
	return NewTransaction(tx, tx, nil)
}
