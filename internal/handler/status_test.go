package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusHandler(t *testing.T) {
	t.Run("удачно", func(t *testing.T) {
		handler := http.HandlerFunc(StatusHandler)
		srv := httptest.NewServer(handler)
		defer srv.Close()

		req := resty.New().R()
		req.Method = http.MethodGet
		req.URL = srv.URL

		resp, err := req.Send()
		require.NoError(t, err)
		assert.Equal(t, "Сервер запущен :)", string(resp.Body()))
	})
}
