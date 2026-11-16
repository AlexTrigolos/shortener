package config

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/config/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestPrepareServer(t *testing.T) {
	t.Run("создали новый Server", func(t *testing.T) {
		preparedServer, err := PrepareServer()
		require.NoError(t, err)

		logger := &zap.Logger{}

		expectedServer := &Server{
			ListenAndServe:   http.ListenAndServe,
			Flags:            Config{Addr: ":8080", BaseURL: "http://localhost:8080", FileStoragePath: ""},
			Logger:           logger,
			LoggerMiddleware: middleware.LoggerMiddleware(logger),
			GzipMiddleware:   middleware.GzipMiddleware,
		}

		require.IsType(t, expectedServer, preparedServer)

		assert.Equal(t, reflect.ValueOf(expectedServer.ListenAndServe), reflect.ValueOf(preparedServer.ListenAndServe))
		assert.Equal(t, expectedServer.Flags, preparedServer.Flags)
		assert.Equal(t, reflect.TypeOf(expectedServer.Logger), reflect.TypeOf(preparedServer.Logger))
		assert.Equal(t, reflect.TypeOf(expectedServer.LoggerMiddleware), reflect.TypeOf(preparedServer.LoggerMiddleware))
		assert.Equal(t, reflect.ValueOf(expectedServer.GzipMiddleware), reflect.ValueOf(preparedServer.GzipMiddleware))
	})
}
