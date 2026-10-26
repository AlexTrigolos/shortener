package config

import (
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrepareServer(t *testing.T) {
	t.Run("создали новый Server", func(t *testing.T) {
		preparedServer, err := PrepareServer()
		require.NoError(t, err)

		expectedServer := &Server{
			ListenAndServe: http.ListenAndServe,
			Flags:          Config{addr: ":8080", baseURL: "http://localhost:8080"},
			Logger:         slog.New(slog.NewTextHandler(os.Stdout, nil)),
		}

		require.IsType(t, expectedServer, preparedServer)

		assert.Equal(t, reflect.ValueOf(expectedServer.ListenAndServe), reflect.ValueOf(preparedServer.ListenAndServe))
		assert.Equal(t, expectedServer.Flags, preparedServer.Flags)
		assert.Equal(t, expectedServer.Logger, preparedServer.Logger)
	})
}
