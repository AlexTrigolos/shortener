package app

import (
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockWithLogger(successLogger *int) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		*successLogger++
		return h
	}
}

func mockSuccessListenAndServe(successLAS *int) func(addr string, handler http.Handler) error {
	return func(addr string, handler http.Handler) error {
		*successLAS++
		return nil
	}
}

func mockPrepareServer(successLogger *int, mockLAS func(addr string, handler http.Handler) error) func() (*config.Server, error) {
	return func() (*config.Server, error) {
		return &config.Server{
			ListenAndServe: mockLAS,
			Flags:          config.Config{},
			Logger:         slog.New(slog.NewTextHandler(os.Stdout, nil)),
			LoggingURI:     mockWithLogger(successLogger),
		}, nil
	}
}

func mockApp(successLogger *int, mockLAS func(addr string, handler http.Handler) error) *App {
	serverFunc := mockPrepareServer(successLogger, mockLAS)
	return &App{prepareServer: serverFunc}
}

func TestNewApp(t *testing.T) {
	t.Run("создать новый App", func(t *testing.T) {
		newApp := NewApp()

		require.IsType(t, &App{prepareServer: config.PrepareServer}, newApp)
	})
}

func TestRunApp(t *testing.T) {
	tests := []struct {
		name               string
		mockListenAndServe func(success *int) func(addr string, handler http.Handler) error
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
			successLAS := 0
			successLogger := 0
			app := mockApp(&successLogger, tt.mockListenAndServe(&successLAS))
			assert.NotPanics(t, app.RunApp)
			if tt.ok {
				assert.Equal(t, 1, successLAS)
				assert.Equal(t, 1, successLogger)
			} else {
				assert.Equal(t, 0, successLAS)
				assert.Equal(t, 0, successLogger)
			}
		})
	}
}
