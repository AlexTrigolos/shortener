package service

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thanhpk/randstr"
)

const baseURL = "http://localhost:8080"

func mockBatchRandStr(executedCount *int, f func(n int, letters ...string) string) func(n int, letters ...string) string {
	return func(n int, letters ...string) string {
		*executedCount++
		return f(n, letters...)
	}
}

func TestNewBatchGenerateURL(t *testing.T) {
	t.Run("создать новый BatchGenerateURL", func(t *testing.T) {
		urls, err := model.NewURL()
		require.NoError(t, err)
		newBatchGenerateRouter := NewBatchGenerateURL(randstr.String, urls)

		require.IsType(t, &BatchGenerateURL{randStr: randstr.String}, newBatchGenerateRouter)
		assert.Equal(t, reflect.ValueOf(GenerateURL{randStr: randstr.String}.randStr), reflect.ValueOf(newBatchGenerateRouter.randStr))
		assert.Equal(t, reflect.ValueOf(GenerateURL{storage: urls}.storage), reflect.ValueOf(newBatchGenerateRouter.storage))
	})
}

func TestBatchGenerateAndSetUniqURL(t *testing.T) {
	originalURLs := []model.BatchRequest{{CorrelationID: "1", OriginalURL: "original_1"}, {CorrelationID: "2", OriginalURL: "original_2"}}
	batchNotUniqURL := strings.Repeat("ABab1221", 10)

	randStr := randstr.String(16)
	goodRandFunc := func(n int, letters ...string) string { return randStr }
	badRandFunc := func(n int, letters ...string) string { return batchNotUniqURL }

	// просто urls
	urls, err := model.NewURL()
	require.NoError(t, err)
	urls.Set(t.Context(), batchNotUniqURL[:8], "value")

	// с файлом
	dirname := t.TempDir()
	defer os.RemoveAll(dirname)

	newFile, err := os.CreateTemp(dirname, "*_new_file.json")
	require.NoError(t, err)
	defer newFile.Close()
	emptyData, err := json.Marshal([]string{})
	require.NoError(t, err)

	_, err = newFile.Write(emptyData)
	require.NoError(t, err)

	urlsWithFile, err := model.NewURLWithFile(newFile.Name())
	require.NoError(t, err)
	urlsWithFile.Set(t.Context(), batchNotUniqURL[:8], "value")

	// БД
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbu, err := model.NewDBURL(db)
	require.NoError(t, err)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO urls \\(short, original\\) VALUES \\(\\$1, \\$2\\)").
		WithArgs(randStr[0:8], "original_1").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO urls \\(short, original\\) VALUES \\(\\$1, \\$2\\)").
		WithArgs(randStr[8:16], "original_2").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	dbBad, mockBad, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	dbuBad, err := model.NewDBURL(dbBad)
	require.NoError(t, err)

	mockBad.ExpectBegin()
	for i := 0; i < 10; i++ {
		mockBad.ExpectExec("INSERT INTO urls \\(short, original\\) VALUES").
			WithArgs("ABab1221", "original_1").
			WillReturnError(errors.New("есть такой short"))
	}
	mockBad.ExpectRollback()
	
	tests := []struct {
		name     string
		mockRand func(n int, letters ...string) string
		storage  BatchSetter
		wantErr  string
	}{
		{
			name:     "URLs создали уникальный урл",
			mockRand: goodRandFunc,
			storage:  urls,
			wantErr:  "",
		},
		{
			name:     "URLs было создано повторение существующего урла",
			mockRand: badRandFunc,
			storage:  urls,
			wantErr:  "исчерпано количество попыток создать короткую ссылку",
		},
		{
			name:     "URLs with file создали уникальный урл",
			mockRand: goodRandFunc,
			storage:  urlsWithFile,
			wantErr:  "",
		},
		{
			name:     "URLs with file было создано повторение существующего урла",
			mockRand: badRandFunc,
			storage:  urlsWithFile,
			wantErr:  "исчерпано количество попыток создать короткую ссылку",
		},
		{
			name:     "DB URLs создали уникальный урл",
			mockRand: goodRandFunc,
			storage:  dbu,
			wantErr:  "",
		},
		{
			name:     "DB URLs было создано повторение существующего урла",
			mockRand: badRandFunc,
			storage:  dbuBad,
			wantErr:  "исчерпано количество попыток создать короткую ссылку",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var executedCount = 0

			newBatchGenerateURL := NewBatchGenerateURL(mockBatchRandStr(&executedCount, tt.mockRand), tt.storage)
			uniqURL, err := newBatchGenerateURL.BatchGenerateAndSetUniqURL(t.Context(), &originalURLs, baseURL)

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Equal(t, &[]model.BatchResponse{}, uniqURL)
				assert.Equal(t, 1, executedCount)
				if _, ok := tt.storage.(*model.DBURL); ok {
					require.NoError(t, mockBad.ExpectationsWereMet())
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, &[]model.BatchResponse{{CorrelationID: "1", ShortURL: baseURL + "/" + randStr[0:8]}, {CorrelationID: "2", ShortURL: baseURL + "/" + randStr[8:16]}}, uniqURL)
				assert.Equal(t, 1, executedCount)
				if _, ok := tt.storage.(*model.DBURL); ok {
					require.NoError(t, mock.ExpectationsWereMet())
				}
			}
		})
	}
}
