package model

import (
	"context"
	"database/sql"
	"errors"
)

type DBURL struct {
	db          *sql.DB
	transaction *sql.Tx
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

func (dbu *DBURL) Begin(_ int) (*Transaction, error) {
	tx, err := dbu.db.Begin()
	if err != nil {
		return nil, err
	}
	dbu.transaction = tx
	transaction, err := NewTransaction(dbu)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (dbu *DBURL) Rollback() error {
	defer func() { dbu.transaction = nil }()
	err := dbu.transaction.Rollback()
	if err != nil {
		return err
	}
	return nil
}

func (dbu *DBURL) Commit() error {
	defer func() { dbu.transaction = nil }()
	err := dbu.transaction.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (dbu *DBURL) SetTx(ctx context.Context, key, value string) error {
	_, err := dbu.transaction.ExecContext(ctx, "INSERT INTO urls (short, original) VALUES ($1, $2)", key, value)
	if err != nil {
		return err
	}
	return nil
}
