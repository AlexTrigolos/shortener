package config

import (
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/auth"
	"github.com/AlexTrigolos/shortener/internal/config/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const secretKey = "key"

func TestPrepareServer(t *testing.T) {
	os.Setenv("SECRET_KEY", secretKey)
	defer func() { os.Unsetenv("SECRET_KEY") }()

	t.Run("создали новый Server", func(t *testing.T) {
		preparedServer, err := PrepareServer()
		require.NoError(t, err)

		logger := &zap.Logger{}
		signedCookie := middleware.NewSignedCookie(auth.NewSigner(secretKey), logger)

		expectedServer := &Server{
			ListenAndServe:         http.ListenAndServe,
			Flags:                  Config{Addr: ":8080", BaseURL: "http://localhost:8080", FileStoragePath: "", SecretKey: secretKey},
			Logger:                 logger,
			LoggerMiddleware:       middleware.LoggerMiddleware(logger),
			GzipMiddleware:         middleware.GzipMiddleware,
			SignedCookieMiddleware: signedCookie.SignedCookieMiddleware,
		}

		require.IsType(t, expectedServer, preparedServer)

		assert.Equal(t, reflect.ValueOf(expectedServer.ListenAndServe), reflect.ValueOf(preparedServer.ListenAndServe))
		assert.Equal(t, expectedServer.Flags, preparedServer.Flags)
		assert.Equal(t, reflect.TypeOf(expectedServer.Logger), reflect.TypeOf(preparedServer.Logger))
		assert.Equal(t, reflect.TypeOf(expectedServer.LoggerMiddleware), reflect.TypeOf(preparedServer.LoggerMiddleware))
		assert.Equal(t, reflect.ValueOf(expectedServer.GzipMiddleware), reflect.ValueOf(preparedServer.GzipMiddleware))
		assert.Equal(t, reflect.TypeOf(expectedServer.SignedCookieMiddleware), reflect.TypeOf(preparedServer.SignedCookieMiddleware))
	})
}
