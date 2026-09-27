package app

import (
	"errors"
	"net/http"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var mockAddr = ""

func mockSuccessNewURL() (*model.URL, error) {
	urls, _ := model.NewURL()
	return urls, nil
}

func mockFailNewURL() (*model.URL, error) {
	urls, _ := model.NewURL()
	return urls, errors.New("не создали urls")
}

func mockSuccessListenAndServe(addr string, handler http.Handler) error {
	mockAddr = addr
	return nil
}

func mockFailListenAndServe(addr string, handler http.Handler) error {
	return errors.New("неудачно настроились и запустились хэндлеры")
}

func TestRun(t *testing.T) {
	tests := []struct {
		name               string
		mockURL            func() (*model.URL, error)
		mockListenAndServe func(addr string, handler http.Handler) error
		err                string
	}{
		{
			name:               "удачный запуск сервера",
			mockURL:            mockSuccessNewURL,
			mockListenAndServe: mockSuccessListenAndServe,
			err:                "",
		},
		{
			name:               "не смогли создать urls",
			mockURL:            mockFailNewURL,
			mockListenAndServe: mockSuccessListenAndServe,
			err:                "не создали urls",
		},
		{
			name:               "удачный запуск сервера",
			mockURL:            mockSuccessNewURL,
			mockListenAndServe: mockFailListenAndServe,
			err:                "неудачно настроились и запустились хэндлеры",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newURL = tt.mockURL
			listenAndServe = tt.mockListenAndServe
			err := Run()

			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, ":8080", mockAddr)
			}
		})
	}
}
