package service

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thanhpk/randstr"
)

const notUniqURL = "ABab1221"

func mockRandStr(executedCount *int, f func(n int, letters ...string) string) func(n int, letters ...string) string {
	return func(n int, letters ...string) string {
		*executedCount++
		return f(n, letters...)
	}
}

func TestNewGenerateURL(t *testing.T) {
	dirname := t.TempDir()
	defer os.RemoveAll(dirname)
	file, err := os.CreateTemp(dirname, "*_file.json")
	require.NoError(t, err)
	filename := file.Name()
	fileData, err := json.Marshal([]string{})
	require.NoError(t, err)
	_, err = file.Write(fileData)
	require.NoError(t, err)

	t.Run("создать новый GenerateURL", func(t *testing.T) {
		urls, err := model.NewURL(filename)
		require.NoError(t, err)
		newGenerateRouter := NewGenerateURL(randstr.String, urls)

		require.IsType(t, &GenerateURL{randStr: randstr.String}, newGenerateRouter)
		assert.Equal(t, reflect.ValueOf(GenerateURL{randStr: randstr.String}.randStr), reflect.ValueOf(newGenerateRouter.randStr))
		assert.Equal(t, reflect.ValueOf(GenerateURL{storage: urls}.storage), reflect.ValueOf(newGenerateRouter.storage))
	})
}

func TestGenerateAndSetUniqURL(t *testing.T) {
	dirname := t.TempDir()
	defer os.RemoveAll(dirname)
	file, err := os.CreateTemp(dirname, "*_file.json")
	require.NoError(t, err)
	filename := file.Name()
	fileData, err := json.Marshal([]string{})
	require.NoError(t, err)
	_, err = file.Write(fileData)
	require.NoError(t, err)

	tests := []struct {
		name     string
		mockRand func(n int, letters ...string) string
		wantErr  string
	}{
		{
			name:     "создали уникальный урл",
			mockRand: randstr.String,
			wantErr:  "",
		},
		{
			name:     "было создано повторение существующего урла",
			mockRand: func(n int, letters ...string) string { return notUniqURL },
			wantErr:  "исчерпано количество попыток создать короткую ссылку",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var executedCount = 0
			urls, err := model.NewURL(filename)
			require.NoError(t, err)
			urls.Set(notUniqURL, "value")

			newGenerateURL := NewGenerateURL(mockRandStr(&executedCount, tt.mockRand), urls)
			uniqURL, err := newGenerateURL.GenerateAndSetUniqURL("originalStr")

			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				assert.Equal(t, "", uniqURL)
				assert.Equal(t, 10, executedCount)
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, notUniqURL, uniqURL)
				assert.Equal(t, 1, executedCount)
				assert.Len(t, []rune(uniqURL), 8)
			}
		})
	}
}
