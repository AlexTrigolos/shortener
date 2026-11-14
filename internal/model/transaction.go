package model

import (
	"context"
	"database/sql"
	"errors"
)

type Executer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type Tx interface {
	Rollback() error
	Commit() error
}

type SelfSetTx interface {
	SetTx(ctx context.Context, key, value string) error
}

type Transaction struct {
	tx        Tx
	executer  Executer
	selfSetTx SelfSetTx
	commited  bool
}

func NewTransaction(tx Tx, executer Executer, selfSetTx SelfSetTx) (*Transaction, error) {
	if executer == nil && selfSetTx == nil {
		return nil, errors.New("должен быть исполнитель запроса установки в транзакции: executer или selfSetTx")
	}
	return &Transaction{tx: tx, executer: executer, selfSetTx: selfSetTx, commited: false}, nil
}

func (t *Transaction) Rollback() error {
	if t.commited {
		return nil
	}
	return t.tx.Rollback()
}

func (t *Transaction) Commit() error {
	t.commited = true
	return t.tx.Commit()
}

func (t *Transaction) SetTx(ctx context.Context, key, value string) (err error) {
	if t.executer != nil {
		_, err = t.executer.ExecContext(ctx, "INSERT INTO urls (short, original) VALUES ($1, $2)", key, value)
	} else if t.selfSetTx != nil {
		err = t.selfSetTx.SetTx(ctx, key, value)
	}
	return
}
