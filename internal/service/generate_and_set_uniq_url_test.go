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

var executedCount = 0

func mockRandStr(n int, letters ...string) string {
	executedCount++
	return notUniqURL
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
		wantErr bool
	}{
		{
			name:    "создали уникальный урл",
			mockRand: randstr.String,
			wantErr: false,
		},
		{
			name:     "было создано повторение существующего урла",
			mockRand: mockRandStr,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls, err := model.NewURL()
			require.NoError(t, err)
			urls.Set(notUniqURL, "value")

			newGenerateURL := NewGenerateURL(tt.mockRand)
			uniqURL, err := newGenerateURL.GenerateAndSetUniqURL(urls, "originalStr")

			if tt.wantErr {
				require.ErrorContains(t, err, "исчерпано количество попыток создать короткую ссылку")
				assert.Equal(t, "", uniqURL)
				assert.Equal(t, 10, executedCount)
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, notUniqURL, uniqURL)
				assert.Len(t, []rune(uniqURL), 8)
			}
		})
	}
}
