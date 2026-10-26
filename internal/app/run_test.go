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

func mockSuccessListenAndServe(success *int) func(addr string, handler http.Handler) error {
	return func(addr string, handler http.Handler) error {
		*success++
		return nil
	}
}

func mockPrepareServer(mockLAS func(addr string, handler http.Handler) error) func() (*config.Server, error) {
	return func() (*config.Server, error) {
		return &config.Server{
			ListenAndServe: mockLAS,
			Flags:          config.Config{},
			Logger:         slog.New(slog.NewTextHandler(os.Stdout, nil)),
		}, nil
	}
}

func mockApp(mockLAS func(addr string, handler http.Handler) error) *App {
	serverFunc := mockPrepareServer(mockLAS)
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
			success := 0
			app := mockApp(tt.mockListenAndServe(&success))
			assert.NotPanics(t, app.RunApp)
			if tt.ok {
				assert.Equal(t, 1, success)
			} else {
				assert.Equal(t, 0, success)
			}
		})
	}
}
