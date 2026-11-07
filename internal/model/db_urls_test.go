package model

import (
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDBURL(t *testing.T) {
	t.Run("удачное получение", func(t *testing.T) {
		db, _, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		urls, err := NewDBURL(db)
		require.NoError(t, err)
		assert.Equal(t, &DBURL{db: db}, urls)
	})
}

func TestDBURL_Set(t *testing.T) {
	tests := []struct {
		name              string
		willSuccessReturn driver.Result
		willFailReturn    error
		wantErr           string
	}{
		{
			name:              "удачное добавление",
			willSuccessReturn: sqlmock.NewResult(1, 1),
			willFailReturn:    nil,
			wantErr:           "",
		},
		{
			name:              "не получилось добавить",
			willSuccessReturn: nil,
			willFailReturn:    errors.New("some db error"),
			wantErr:           "some db error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			dbu := &DBURL{db: db}

			ctx := t.Context()
			key := "abc"
			value := "https://google.com"

			if tt.wantErr == "" {
				mock.ExpectExec("INSERT INTO urls \\(short, original\\) VALUES").
					WithArgs(key, value).
					WillReturnResult(tt.willSuccessReturn)

				require.NoError(t, dbu.Set(ctx, key, value))
			} else {
				mock.ExpectExec("INSERT INTO urls \\(short, original\\) VALUES").
					WithArgs(key, value).
					WillReturnError(tt.willFailReturn)

				require.ErrorContains(t, dbu.Set(ctx, key, value), tt.wantErr)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBURL_Get(t *testing.T) {
	tests := []struct {
		name              string
		willSuccessReturn *sqlmock.Rows
		willFailReturn    error
		wantErr           string
	}{
		{
			name:              "удачное получение",
			willSuccessReturn: sqlmock.NewRows([]string{"original"}).AddRow("url_value"),
			willFailReturn:    nil,
			wantErr:           "",
		},
		{
			name:              "не удалось получить",
			willSuccessReturn: nil,
			willFailReturn:    errors.New("some db error"),
			wantErr:           "some db error",
		},
		{
			name:              "получили null",
			willSuccessReturn: sqlmock.NewRows([]string{"original"}).AddRow(nil),
			willFailReturn:    nil,
			wantErr:           "такой ключ не найден",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			dbu := &DBURL{db: db}

			ctx := t.Context()
			key := "abc"

			if tt.willFailReturn == nil {
				mock.ExpectQuery("SELECT original FROM urls WHERE short = \\$1").
					WithArgs(key).
					WillReturnRows(tt.willSuccessReturn)
			} else {
				mock.ExpectQuery("SELECT original FROM urls WHERE short = \\$1").
					WithArgs(key).
					WillReturnError(tt.willFailReturn)
			}

			value, err := dbu.Get(ctx, key)
			if tt.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, "url_value", value)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Equal(t, "", value)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBURL_Begin(t *testing.T) {
	t.Run("удачное создание транзакции", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		dbu, err := NewDBURL(db)
		require.NoError(t, err)

		mock.ExpectBegin()
		tr, err := dbu.Begin(0)
		require.NoError(t, err)
		assert.Equal(t, &Transaction{tx: dbu, commited: false}, tr)
		assert.NotNil(t, dbu.transaction)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDBURL_Rollback(t *testing.T) {
	t.Run("удачная отмена транзакции", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		dbu, err := NewDBURL(db)
		require.NoError(t, err)
		mock.ExpectBegin()
		_, err = dbu.Begin(0)
		require.NoError(t, err)

		mock.ExpectRollback()
		err = dbu.Rollback()
		require.NoError(t, err)
		assert.Nil(t, dbu.transaction)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDBURL_Commit(t *testing.T) {
	t.Run("удачное завершение транзакции", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		dbu, err := NewDBURL(db)
		require.NoError(t, err)
		mock.ExpectBegin()
		_, err = dbu.Begin(0)
		require.NoError(t, err)

		mock.ExpectCommit()
		err = dbu.Commit()
		require.NoError(t, err)
		assert.Nil(t, dbu.transaction)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDBURL_SetTx(t *testing.T) {
	t.Run("удачное добавление внутри транзакции", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		dbu, err := NewDBURL(db)
		require.NoError(t, err)
		mock.ExpectBegin()
		_, err = dbu.Begin(0)
		require.NoError(t, err)

		mock.ExpectExec("INSERT INTO urls \\(short, original\\) VALUES \\(\\$1, \\$2\\)").
			WithArgs("key_1", "value_1").
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO urls \\(short, original\\) VALUES \\(\\$1, \\$2\\)").
			WithArgs("key_2", "value_2").
			WillReturnResult(sqlmock.NewResult(2, 1))
		err = dbu.SetTx(t.Context(), "key_1", "value_1")
		require.NoError(t, err)
		err = dbu.SetTx(t.Context(), "key_2", "value_2")
		require.NoError(t, err)
		assert.NotNil(t, dbu.transaction)
		require.NoError(t, mock.ExpectationsWereMet())
	})
	t.Run("неудачное добавление внутри транзакции", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		dbu, err := NewDBURL(db)
		require.NoError(t, err)
		mock.ExpectBegin()
		_, err = dbu.Begin(0)
		require.NoError(t, err)

		mock.ExpectExec("INSERT INTO urls \\(short, original\\) VALUES \\(\\$1, \\$2\\)").
			WithArgs("key_1", "value_1").
			WillReturnError(errors.New("не смогли добавить"))
		err = dbu.SetTx(t.Context(), "key_1", "value_1")
		require.ErrorContains(t, err, "не смогли добавить")
		assert.NotNil(t, dbu.transaction)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
