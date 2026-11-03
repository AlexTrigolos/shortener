package app

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func mockLoggerMiddleware(successLogger *int) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		*successLogger++
		return h
	}
}

func mockGzipMiddleware(successGzip *int) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		*successGzip++
		return h
	}
}

func mockSuccessListenAndServe(successLAS *int) func(addr string, handler http.Handler) error {
	return func(addr string, handler http.Handler) error {
		*successLAS++
		return nil
	}
}

func mockPrepareServer(successLogger *int, successGzip *int, filename string, successLAS *int, logger *zap.Logger, err error) func() (*config.Server, error) {
	return func() (*config.Server, error) {
		return &config.Server{
			ListenAndServe:   mockSuccessListenAndServe(successLAS),
			Flags:            config.Config{FileStoragePath: filename},
			Logger:           logger,
			LoggerMiddleware: mockLoggerMiddleware(successLogger),
			GzipMiddleware:   mockGzipMiddleware(successGzip),
		}, err
	}
}

func mockApp(successLogger *int, successGzip *int, filename string, successLAS *int, logger *zap.Logger, err error) *App {
	serverFunc := mockPrepareServer(successLogger, successGzip, filename, successLAS, logger, err)
	return &App{prepareServer: serverFunc}
}

func TestNewApp(t *testing.T) {
	t.Run("создать новый App", func(t *testing.T) {
		newApp := NewApp()

		require.IsType(t, &App{prepareServer: config.PrepareServer}, newApp)
	})
}

func TestRunApp(t *testing.T) {
	dirname := t.TempDir()
	defer os.RemoveAll(dirname)
	file, err := os.CreateTemp(dirname, "*_file.json")
	require.NoError(t, err)
	defer file.Close()
	filename := file.Name()
	fileData, err := json.Marshal([]string{})
	require.NoError(t, err)
	_, err = file.Write(fileData)
	require.NoError(t, err)

	tests := []struct {
		name               string
		logger             *zap.Logger
		serverErr          error
		errMessage         string
	}{
		{
			name:               "удачный запуск",
			logger:             zap.NewNop(),
			serverErr:          nil,
			errMessage:         "",
		},
		{
			name:               "ошибка создания логера",
			logger:             zap.NewNop(),
			serverErr:          errors.New("что-то сломалось"),
			errMessage:         "что-то сломалось",
		},
		{
			name:               "ошибка подготовки сервера",
			logger:             nil,
			serverErr:          errors.New("не удалось создать логгер"),
			errMessage:         "не удалось создать логгер",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			successLAS, successLogger, successGzip := 0, 0, 0
			app := mockApp(&successLogger, &successGzip, filename, &successLAS, tt.logger, tt.serverErr)
			err := app.RunApp()
			if tt.errMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, 1, successLAS)
				assert.Equal(t, 1, successLogger)
			} else {
				require.ErrorContains(t, err, tt.errMessage)
				assert.Equal(t, 0, successLAS)
				assert.Equal(t, 0, successLogger)
			}
		})
	}
}
