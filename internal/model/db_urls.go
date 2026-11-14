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

func (dbu *DBURL) Set(ctx context.Context, key, value string) error {
	_, err := dbu.db.ExecContext(ctx, "INSERT INTO urls (short, original) VALUES ($1, $2)", key, value)
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

func (dbu *DBURL) GetByOriginal(ctx context.Context, original string) (string, error) {
	row := dbu.db.QueryRowContext(ctx, "SELECT short FROM urls WHERE original = $1", original)
	var short string

	err := row.Scan(&short)
	if err != nil {
		return "", err
	}
	return short, nil
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

func (dbu *DBURL) Close() error {
	return dbu.db.Close()
}
