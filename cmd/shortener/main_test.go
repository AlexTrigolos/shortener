package main

import (
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/config"
	"github.com/stretchr/testify/assert"
)

var success int = 0

func mockSuccessListenAndServe(addr string, handler http.Handler) error {
	success++
	return nil
}

func mockPrepareServer(mockLAS func(addr string, handler http.Handler) error) func() *config.Server {
	return func() *config.Server {
		return &config.Server{
			ListenAndServe: mockLAS,
			Flags:          config.Config{},
			Logger:         slog.New(slog.NewTextHandler(os.Stdout, nil)),
		}
	}
}

func TestMain(t *testing.T) {
	tests := []struct {
		name               string
		mockListenAndServe func(addr string, handler http.Handler) error
		ok                 bool
	}{
		{
			name:               "удачный запуск",
			mockListenAndServe: mockSuccessListenAndServe,
			ok:                 true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			success = 0
			oldPrepareServer := prepareServer
			prepareServer = mockPrepareServer(tt.mockListenAndServe)
			defer func() { prepareServer = oldPrepareServer }()
			assert.NotPanics(t, main)
			// неудачный запуск не пропускается вторым интеграционным тестом, там ожидается, что ошибок не будет
			if tt.ok {
				assert.Equal(t, 1, success)
			}
		})
	}
}
