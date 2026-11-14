package model

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockMigration struct{}

func (*MockMigration) RunMigrations(db *sql.DB) (*database.Driver, *migrate.Migrate, error) {
	return nil, nil, nil
}

func TestNewDBURL(t *testing.T) {
	t.Run("удачное получение", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		// mock.ExpectQuery(".+")...
		mock.ExpectQuery(".+").WillReturnRows(sqlmock.NewRows([]string{"dummy"}).AddRow("val"))
		mock.ExpectExec(".+").WillReturnResult(sqlmock.NewResult(0, 0))

		urls, _, _, err := NewDBURL(db, &MockMigration{})
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
		{
			name:              "ошибка добавления pgconn.PgError",
			willSuccessReturn: nil,
			willFailReturn:    &pgconn.PgError{Code: pgerrcode.AdminShutdown, Message: "отключка"},
			wantErr:           ": отключка (SQLSTATE 57P01)",
		},
		{
			name:              "ошибка добавления конфликт",
			willSuccessReturn: nil,
			willFailReturn:    &pgconn.PgError{Code: pgerrcode.UniqueViolation, Message: "уникальности ошибку ты получил"},
			wantErr:           "уникальности ошибку ты получил",
		},
		{
			name:              "ошибка добавления конфликт значений",
			willSuccessReturn: nil,
			willFailReturn:    &pgconn.PgError{Code: pgerrcode.UniqueViolation, Message: "уникальности ошибку ты получил", ConstraintName: "idx_urls_original"},
			wantErr:           "conflict",
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

func TestDBURL_GetByOriginal(t *testing.T) {
	tests := []struct {
		name              string
		willSuccessReturn *sqlmock.Rows
		willFailReturn    error
		wantErr           string
	}{
		{
			name:              "удачное получение",
			willSuccessReturn: sqlmock.NewRows([]string{"short"}).AddRow("shory_url"),
			willFailReturn:    nil,
			wantErr:           "",
		},
		{
			name:              "не удалось получить",
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
			value := "original_value"

			if tt.willFailReturn == nil {
				mock.ExpectQuery("SELECT short FROM urls WHERE original = \\$1").
					WithArgs(value).
					WillReturnRows(tt.willSuccessReturn)
			} else {
				mock.ExpectQuery("SELECT short FROM urls WHERE original = \\$1").
					WithArgs(value).
					WillReturnError(tt.willFailReturn)
			}

			key, err := dbu.GetByOriginal(ctx, value)
			if tt.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, "shory_url", key)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Equal(t, "", key)
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

		dbu, _, _, err := NewDBURL(db, &MockMigration{})
		require.NoError(t, err)

		mock.ExpectBegin()
		tr, err := dbu.Begin(0)
		require.NoError(t, err)
		assert.Equal(t, &Transaction{tx: tr.tx, executer: tr.executer, selfSetTx: nil, commited: false}, tr)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
