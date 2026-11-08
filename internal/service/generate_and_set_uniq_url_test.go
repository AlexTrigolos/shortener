package service

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thanhpk/randstr"
)

const (
	notUniqURL      = "ABab1221"
	notUniqOriginal = "value"
)

func mockRandStr(executedCount *int, f func(n int, letters ...string) string) func(n int, letters ...string) string {
	return func(n int, letters ...string) string {
		*executedCount++
		return f(n, letters...)
	}
}

func TestNewGenerateURL(t *testing.T) {
	t.Run("создать новый GenerateURL", func(t *testing.T) {
		urls, err := model.NewURL()
		require.NoError(t, err)
		newGenerateRouter := NewGenerateURL(randstr.String, urls)

		require.IsType(t, &GenerateURL{randStr: randstr.String}, newGenerateRouter)
		assert.Equal(t, reflect.ValueOf(GenerateURL{randStr: randstr.String}.randStr), reflect.ValueOf(newGenerateRouter.randStr))
		assert.Equal(t, reflect.ValueOf(GenerateURL{storage: urls}.storage), reflect.ValueOf(newGenerateRouter.storage))
	})
}

func TestGenerateAndSetUniqURL(t *testing.T) {
	randStr := randstr.String(8)
	goodRandFunc := func(n int, letters ...string) string { return randStr }
	badRandFunc := func(n int, letters ...string) string { return notUniqURL }

	// просто urls
	urls, err := model.NewURL()
	require.NoError(t, err)
	urls.Set(t.Context(), notUniqURL, notUniqOriginal)

	// с файлом
	dirname := t.TempDir()
	defer os.RemoveAll(dirname)

	file, err := os.CreateTemp(dirname, "*_file.json")
	require.NoError(t, err)
	defer file.Close()
	emptyData, err := json.Marshal([]string{})
	require.NoError(t, err)

	_, err = file.Write(emptyData)
	require.NoError(t, err)

	urlsWithFile, err := model.NewURLWithFile(file.Name())
	require.NoError(t, err)
	urlsWithFile.Set(t.Context(), notUniqURL, notUniqOriginal)

	// БД
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbu, err := model.NewDBURL(db)
	require.NoError(t, err)

	mock.ExpectExec("INSERT INTO urls \\(short, original\\) VALUES \\(\\$1, \\$2\\)").
		WithArgs(randStr, "originalStr_1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	dbBad, mockBad, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbuBad, err := model.NewDBURL(dbBad)
	require.NoError(t, err)
	for i := 0; i < 10; i++ {
		mockBad.ExpectExec("INSERT INTO urls \\(short, original\\) VALUES").
			WithArgs("ABab1221", "originalStr_2").
			WillReturnError(errors.New("исчерпано количество попыток создать короткую ссылку"))
	}

	dbConflict, mockConflict, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbuConflict, err := model.NewDBURL(dbConflict)
	require.NoError(t, err)

	conflictRandStr := randstr.String(8)
	ConflictRandFunc := func(n int, letters ...string) string { return conflictRandStr }
	mockConflict.ExpectExec("INSERT INTO urls \\(short, original\\) VALUES \\(\\$1, \\$2\\)").
		WithArgs(conflictRandStr, notUniqOriginal).
		WillReturnError(&pgconn.PgError{Code: pgerrcode.UniqueViolation, Message: "duplicate key value violates unique constraint"})
	mockConflict.ExpectQuery("SELECT short FROM urls WHERE original = \\$1").
		WithArgs(notUniqOriginal).
		WillReturnRows(sqlmock.NewRows([]string{"short"}).AddRow(notUniqURL))

	type want struct {
		err          string
		result       string
		executeCount int
	}
	tests := []struct {
		name        string
		mockRand    func(n int, letters ...string) string
		storage     SetterWithConflict
		newOriginal string
		want        want
	}{
		{
			name:        "URLs создали уникальный урл",
			mockRand:    goodRandFunc,
			storage:     urls,
			newOriginal: "originalStr_1",
			want: want{
				err:          "",
				result:       randStr,
				executeCount: 1,
			},
		},
		{
			name:        "URLs было создано повторение существующего урла",
			mockRand:    badRandFunc,
			storage:     urls,
			newOriginal: "originalStr_2",
			want: want{
				err:          "исчерпано количество попыток создать короткую ссылку",
				result:       "",
				executeCount: 10,
			},
		},
		{
			name:        "URLs добавляемое значение не уникально",
			mockRand:    goodRandFunc,
			storage:     urls,
			newOriginal: notUniqOriginal,
			want: want{
				err:          "conflict",
				result:       notUniqURL,
				executeCount: 1,
			},
		},
		{
			name:        "URLsWithFile создали уникальный урл",
			mockRand:    goodRandFunc,
			storage:     urlsWithFile,
			newOriginal: "originalStr_1",
			want: want{
				err:          "",
				result:       randStr,
				executeCount: 1,
			},
		},
		{
			name:        "URLsWithFile было создано повторение существующего урла",
			mockRand:    badRandFunc,
			storage:     urlsWithFile,
			newOriginal: "originalStr_2",
			want: want{
				err:          "исчерпано количество попыток создать короткую ссылку",
				result:       "",
				executeCount: 10,
			},
		},
		{
			name:        "URLsWithFile добавляемое значение не уникально",
			mockRand:    goodRandFunc,
			storage:     urlsWithFile,
			newOriginal: notUniqOriginal,
			want: want{
				err:          "conflict",
				result:       notUniqURL,
				executeCount: 1,
			},
		},
		{
			name:        "DBURLs создали уникальный урл",
			mockRand:    goodRandFunc,
			storage:     dbu,
			newOriginal: "originalStr_1",
			want: want{
				err:          "",
				result:       randStr,
				executeCount: 1,
			},
		},
		{
			name:        "DBURLs было создано повторение существующего урла",
			mockRand:    badRandFunc,
			storage:     dbuBad,
			newOriginal: "originalStr_2",
			want: want{
				err:          "исчерпано количество попыток создать короткую ссылку",
				result:       "",
				executeCount: 10,
			},
		},
		{
			name:        "DBURLs добавляемое значение не уникально",
			mockRand:    ConflictRandFunc,
			storage:     dbuConflict,
			newOriginal: notUniqOriginal,
			want: want{
				err:          "conflict",
				result:       notUniqURL,
				executeCount: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var executedCount = 0

			newGenerateURL := NewGenerateURL(mockRandStr(&executedCount, tt.mockRand), tt.storage)
			uniqURL, err := newGenerateURL.GenerateAndSetUniqURL(t.Context(), tt.newOriginal)

			if tt.want.err != "" {
				require.ErrorContains(t, err, tt.want.err)
				if _, ok := tt.storage.(*model.DBURL); ok {
					if tt.want.err == "conflict" {
						require.NoError(t, mockConflict.ExpectationsWereMet())
					} else {
						require.NoError(t, mockBad.ExpectationsWereMet())
					}
				}
			} else {
				require.NoError(t, err)
				assert.Len(t, []rune(uniqURL), 8)
				if _, ok := tt.storage.(*model.DBURL); ok {
					require.NoError(t, mock.ExpectationsWereMet())
				}
			}
			assert.Equal(t, tt.want.result, uniqURL)
			assert.Equal(t, tt.want.executeCount, executedCount)
		})
	}
}
