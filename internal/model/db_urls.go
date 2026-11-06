package model

import (
	"context"
	"database/sql"
	"errors"
)

type DBURL struct {
	db *sql.DB
}

func NewDBURL(db *sql.DB) (*DBURL, error) {
	return &DBURL{db: db}, nil
}

func (dbu *DBURL) Set(ctx context.Context, key, value string) error {
	_, err := dbu.db.ExecContext(ctx, "INSERT INTO urls (short, original) VALUES ($1, $2)", key, value)
	if err != nil {
		return err
	}
	return nil
}

func (dbu *DBURL) Get(ctx context.Context, key string) (string, error) {
	row := dbu.db.QueryRowContext(ctx, "SELECT original FROM urls WHERE short = $1", key)
	var original sql.NullString

	err := row.Scan(&original)
	if err != nil {
		return "", err
	}
	if original.Valid {
		return original.String, nil
	}

	return "", errors.New("такой ключ не найден")
}

func (dbu *DBURL) PingContext(ctx context.Context) error {
	return dbu.db.PingContext(ctx)
}
