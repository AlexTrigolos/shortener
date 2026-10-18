package service

import (
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

func mockNewURL() (*struct{}, error) {
	return &struct{}{}, nil
}

func TestNewGenerateURL(t *testing.T) {
	t.Run("создать новый GenerateURL", func(t *testing.T) {
		newGenerateRouter := NewGenerateURL(randstr.String)

		require.IsType(t, &GenerateURL{randStr: randstr.String}, newGenerateRouter)
		assert.Equal(t, reflect.ValueOf(GenerateURL{randStr: randstr.String}.randStr), reflect.ValueOf(newGenerateRouter.randStr))
	})
}

func TestGenerateAndSetUniqURL(t *testing.T) {
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
		{
			name:     "передан интерфейс не реализующий Setter",
			mockRand: randstr.String,
			wantErr:  "не был реализован интерфейс для генерации короткой ссылки",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var executedCount = 0
			// отдельно обрабатываю ошибку интерфейса, потому что не удается нормально замокать создание псевдо URL без Setter'а
			if tt.wantErr == "не был реализован интерфейс для генерации короткой ссылки" {
				fakeURLs, err := mockNewURL()
				require.NoError(t, err)

				newGenerateURL := NewGenerateURL(mockRandStr(&executedCount, tt.mockRand))
				uniqURL, err := newGenerateURL.GenerateAndSetUniqURL(fakeURLs, "originalStr")

				require.ErrorContains(t, err, tt.wantErr)
				assert.Equal(t, "", uniqURL)
				assert.Equal(t, 0, executedCount)
			} else {
				urls, err := model.NewURL()
				require.NoError(t, err)
				urls.Set(notUniqURL, "value")

				newGenerateURL := NewGenerateURL(mockRandStr(&executedCount, tt.mockRand))
				uniqURL, err := newGenerateURL.GenerateAndSetUniqURL(urls, "originalStr")

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
			}
		})
	}
}
