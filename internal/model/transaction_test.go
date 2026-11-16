package model

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockTx struct{}

func (*MockTx) Rollback() error {
	return errors.New("error")
}
func (*MockTx) Commit() error {
	return nil
}

type MockExecuter struct{}

func (*MockExecuter) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return nil, nil
}

type MockTxSetter struct{}

func (*MockTxSetter) SetTx(ctx context.Context, key, value string) error {
	return nil
}

func TestNewTransaction(t *testing.T) {
	t.Run("удачное получение", func(t *testing.T) {
		newTransaction, err := NewTransaction(&MockTx{}, &MockExecuter{}, &MockTxSetter{})
		require.NoError(t, err)
		assert.Equal(t, &Transaction{tx: &MockTx{}, executer: &MockExecuter{}, selfSetTx: &MockTxSetter{}, commited: false}, newTransaction)
	})
	t.Run("нет установщика значения в транзакции", func(t *testing.T) {
		newTransaction, err := NewTransaction(&MockTx{}, nil, nil)
		require.Error(t, err, "должен быть исполнитель запроса установки в транзакции: executer или selfSetTx")
		assert.Nil(t, newTransaction)
	})
}

func TestTransaction_Rollback(t *testing.T) {
	notCommited, err := NewTransaction(&MockTx{}, &MockExecuter{}, &MockTxSetter{})
	require.NoError(t, err)

	commited, err := NewTransaction(&MockTx{}, &MockExecuter{}, &MockTxSetter{})
	require.NoError(t, err)
	err = commited.Commit()
	require.NoError(t, err)

	tests := []struct {
		name        string
		transaction *Transaction
		wantExecute bool
	}{
		{
			name:        "удачно откатились",
			transaction: notCommited,
			wantExecute: true,
		},
		{
			name:        "был коммит",
			transaction: commited,
			wantExecute: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.transaction.Rollback()
			// для тестирования замоканная транзакция (не модель) кидает ошибку для Rollback, то есть при вызове отката транзакции будет ошибка
			if tt.wantExecute {
				require.ErrorContains(t, err, "error")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTransaction_Commit(t *testing.T) {
	t.Run("вызывается", func(t *testing.T) {
		tr, err := NewTransaction(&MockTx{}, &MockExecuter{}, &MockTxSetter{})
		require.NoError(t, err)
		err = tr.Commit()
		require.NoError(t, err)
		assert.Equal(t, true, tr.commited)
	})
}

func TestTransaction_SetTx(t *testing.T) {
	urls, err := NewURL()
	require.NoError(t, err)
	urlsTR, err := urls.Begin(0)
	require.NoError(t, err)

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()
	
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO urls \\(short, original\\) VALUES").
		WithArgs("key", "value").
		WillReturnResult(sqlmock.NewResult(1, 1))

	dbTR, err := db.Begin()
	require.NoError(t, err)
	tests := []struct {
		name        string
		tx          Tx
		executer    Executer
		selfSetTx   SelfSetTx
		wantDBMock  bool
	}{
		{
			name:        "выполняется для бд хранилища",
			tx:          &MockTx{},
			executer:    dbTR,
			selfSetTx:   nil,
			wantDBMock:  true,
		},
		{
			name:        "выполняется для хранилища в памяти",
			executer:    nil,
			tx:          &MockTx{},
			selfSetTx:   urlsTR,
			wantDBMock:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transaction, err := NewTransaction(tt.tx, tt.executer, tt.selfSetTx)
			require.NoError(t, err)
			err = transaction.SetTx(t.Context(), "key", "value")
			require.NoError(t, err)
			if tt.wantDBMock {
				require.NoError(t, mock.ExpectationsWereMet())
			} else {
				urlsTR.Commit()
				value, err := urls.Get(t.Context(), "key")
				require.NoError(t, err)
				assert.Equal(t, value, "value")
			}
		})
	}
}
