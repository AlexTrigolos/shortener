package config

import (
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/config/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareServer(t *testing.T) {
	t.Run("создали новый Server", func(t *testing.T) {
		preparedServer, err := PrepareServer()
		require.NoError(t, err)

		expectedServer := &Server{
			ListenAndServe:   http.ListenAndServe,
			Flags:            Config{Addr: ":8080", BaseURL: "http://localhost:8080", FileStoragePath: "./urls_storage.json"},
			Logger:           slog.New(slog.NewTextHandler(os.Stdout, nil)),
			LoggerMiddleware: middleware.LoggerMiddleware,
			GzipMiddleware:   middleware.GzipMiddleware,
		}

		require.IsType(t, expectedServer, preparedServer)

		assert.Equal(t, reflect.ValueOf(expectedServer.ListenAndServe), reflect.ValueOf(preparedServer.ListenAndServe))
		assert.Equal(t, expectedServer.Flags, preparedServer.Flags)
		assert.Equal(t, expectedServer.Logger, preparedServer.Logger)
		assert.Equal(t, reflect.ValueOf(expectedServer.LoggerMiddleware), reflect.ValueOf(preparedServer.LoggerMiddleware))
		assert.Equal(t, reflect.ValueOf(expectedServer.GzipMiddleware), reflect.ValueOf(preparedServer.GzipMiddleware))
	})
}
