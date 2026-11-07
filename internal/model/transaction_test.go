package model

import (
	"errors"
	"testing"

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

func TestNewTransaction(t *testing.T) {
	t.Run("удачное получение", func(t *testing.T) {
		newTransaction, err := NewTransaction(&MockTx{})
		require.NoError(t, err)
		assert.Equal(t, &Transaction{tx: &MockTx{}, commited: false}, newTransaction)
	})
}

func TestTransaction_Rollback(t *testing.T) {
	notCommited, err := NewTransaction(&MockTx{})
	require.NoError(t, err)

	commited, err := NewTransaction(&MockTx{})
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
			if tt.wantExecute{
				require.ErrorContains(t, err, "error")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTransaction_Commit(t *testing.T) {
	t.Run("вызывается", func(t *testing.T) {
		tr, err := NewTransaction(&MockTx{})
		require.NoError(t, err)
		err = tr.Commit()
		require.NoError(t, err)
		assert.Equal(t, true, tr.commited)
	})
}
