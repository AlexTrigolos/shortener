package service

import (
	"testing"

	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thanhpk/randstr"
)

const notUniqURL = "ABab1221"

var (
	executedMock  = false
	executedCount = 0
)

func mockRandStr(n int, letters ...string) string {
	executedCount++
	if !executedMock {
		executedMock = true
		return notUniqURL
	} else {
		return randstr.String(8)
	}
}

func TestGenerateUniqURL(t *testing.T) {
	urls, err := model.NewURL()
	require.NoError(t, err)
	urls.Set(notUniqURL, "value")

	tests := []struct {
		name     string
		mockRand func(n int, letters ...string) string
		notWant  string
	}{
		{
			name:    "создали уникальный урл",
			notWant: notUniqURL,
		},
		{
			name:     "было создано повторение существующего урла",
			mockRand: mockRandStr,
			notWant:  notUniqURL,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.mockRand != nil {
				randStr = tt.mockRand
			}
			uniqURL := GenerateUniqURL(urls)
			assert.NotEqual(t, tt.notWant, uniqURL)
			assert.Len(t, []rune(uniqURL), 8)
			if tt.mockRand != nil {
				assert.Equal(t, true, executedMock)
				assert.Equal(t, 2, executedCount)
			}
		})
	}
}
