package model

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/lib/pq"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	dbUserID  = "user_id"
	dbBaseURL = "http://base_url.org"
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
			willFailReturn:    &pgconn.PgError{Code: pgerrcode.UniqueViolation, Message: "уникальности ошибку ты получил", ConstraintName: "uniq_original_not_deleted"},
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
				mock.ExpectExec("INSERT INTO urls \\(short, original, user_id\\) VALUES").
					WithArgs(key, value, dbUserID).
					WillReturnResult(tt.willSuccessReturn)

				require.NoError(t, dbu.Set(ctx, key, value, dbUserID))
			} else {
				mock.ExpectExec("INSERT INTO urls \\(short, original, user_id\\) VALUES").
					WithArgs(key, value, dbUserID).
					WillReturnError(tt.willFailReturn)

				require.ErrorContains(t, dbu.Set(ctx, key, value, dbUserID), tt.wantErr)
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
			willSuccessReturn: sqlmock.NewRows([]string{"original", "delete_flag"}).AddRow("url_value", "false"),
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
			name:              "удален",
			willSuccessReturn: sqlmock.NewRows([]string{"original", "delete_flag"}).AddRow("url_value", "true"),
			willFailReturn:    nil,
			wantErr:           "deleted",
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
				mock.ExpectQuery("SELECT original, delete_flag FROM urls WHERE short = \\$1").
					WithArgs(key).
					WillReturnRows(tt.willSuccessReturn)
			} else {
				mock.ExpectQuery("SELECT original, delete_flag FROM urls WHERE short = \\$1").
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
			willSuccessReturn: sqlmock.NewRows([]string{"short"}).AddRow("short_url"),
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
				mock.ExpectQuery("SELECT short FROM urls WHERE original = \\$1 AND delete_flag = FALSE").
					WithArgs(value).
					WillReturnRows(tt.willSuccessReturn)
			} else {
				mock.ExpectQuery("SELECT short FROM urls WHERE original = \\$1 AND delete_flag = FALSE").
					WithArgs(value).
					WillReturnError(tt.willFailReturn)
			}

			key, err := dbu.GetByOriginal(ctx, value, dbBaseURL)
			if tt.wantErr == "" {
				require.NoError(t, err)
				assert.Equal(t, dbBaseURL+"/short_url", key)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Equal(t, "", key)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBURL_GetByUserID(t *testing.T) {
	tests := []struct {
		name              string
		willSuccessReturn *sqlmock.Rows
		willFailReturn    error
		wantErr           string
	}{
		{
			name:              "удачное получение",
			willSuccessReturn: sqlmock.NewRows([]string{"short", "original"}).AddRow("short_url", "origonal_url").AddRow("short_url2", "origonal_url2"),
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
			userID := "user_id"

			if tt.willFailReturn == nil {
				mock.ExpectQuery("SELECT short, original FROM urls WHERE user_id = \\$1 AND delete_flag = FALSE").
					WithArgs(userID).
					WillReturnRows(tt.willSuccessReturn)
			} else {
				mock.ExpectQuery("SELECT short, original FROM urls WHERE user_id = \\$1 AND delete_flag = FALSE").
					WithArgs(userID).
					WillReturnError(tt.willFailReturn)
			}

			userURL, err := dbu.GetByUserID(ctx, userID, dbBaseURL)
			if tt.wantErr == "" {
				require.NoError(t, err)
				expected := &[]UserURL{{ShortURL: dbBaseURL + "/short_url", OriginalURL: "origonal_url"}, {ShortURL: dbBaseURL + "/short_url2", OriginalURL: "origonal_url2"}}
				assert.Equal(t, expected, userURL)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Nil(t, userURL)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBURL_CheckUserIDURLs(t *testing.T) {
	tests := []struct {
		name      string
		resShorts int
		wantErr   string
	}{
		{
			name:      "удачная проверка урлов пользователя",
			resShorts: 2,
			wantErr:   "",
		},
		{
			name:      "не все урлы его",
			resShorts: 0,
			wantErr:   "не все переданные urls принадлежат пользователю",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			dbu := &DBURL{db: db}

			urls := []string{"short_url", "short_url_2"}
			mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM urls WHERE user_id = \\$1 AND short = ANY\\(\\$2\\)").
				WithArgs(dbUserID, pq.Array(urls)).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(tt.resShorts))
			err = dbu.CheckUserIDURLs(t.Context(), urls, dbUserID)
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.ErrorContains(t, err, tt.wantErr)
			}
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDBURL_DeleteURLs(t *testing.T) {
	t.Run("удачное удаление", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		require.NoError(t, err)
		defer db.Close()

		dbu := &DBURL{db: db}

		urls := []URLsData{{Urls: []string{"short_url"}, UserID: "user_id"}, {Urls: []string{"short_url_2"}, UserID: "user_id_2"}}
		mock.ExpectExec("UPDATE urls SET delete_flag = TRUE WHERE short = ANY\\(\\$1\\)").
			WithArgs(pq.Array([]string{"short_url", "short_url_2"})).
			WillReturnResult(sqlmock.NewResult(0, 0))
		err = dbu.DeleteURLs(t.Context(), urls)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
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
