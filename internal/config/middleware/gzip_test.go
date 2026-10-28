package middleware

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/AlexTrigolos/shortener/internal/handler"
	"github.com/AlexTrigolos/shortener/internal/handler/api"
	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/AlexTrigolos/shortener/internal/service"
)

func mockRandStr(n int, letters ...string) string {
	return "short"
}

func TestGzipMiddleware(t *testing.T) {
	t.Run("sends_gzip", func(t *testing.T) {
		urls, err := model.NewURL()
		require.NoError(t, err)
		handler := GzipMiddleware(http.HandlerFunc(api.NewCompact(service.NewGenerateURL(mockRandStr, urls)).CompactHandler("https://host:port")))
		srv := httptest.NewServer(handler)
		defer srv.Close()
		requestBody := `{"url": "https://google.com"}`
		successBody := `{"result": "https://host:port/short"}`

		buf := bytes.NewBuffer(nil)
		zb := gzip.NewWriter(buf)
		_, err = zb.Write([]byte(requestBody))
		require.NoError(t, err)
		err = zb.Close()
		require.NoError(t, err)

		r := httptest.NewRequest("POST", srv.URL+"/api/shorten", buf)
		r.RequestURI = ""
		r.Header.Set("Content-Encoding", "gzip")
		r.Header.Set("Content-Type", "application/json")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)

		defer resp.Body.Close()

		b, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.JSONEq(t, successBody, string(b))
	})

	t.Run("accepts_gzip", func(t *testing.T) {
		handler := GzipMiddleware(http.HandlerFunc(handler.StatusHandler))
		srv := httptest.NewServer(handler)
		defer srv.Close()
		successBody := `Server started :)`

		buf := bytes.NewBufferString("")
		r := httptest.NewRequest("GET", srv.URL, buf)
		r.RequestURI = ""
		r.Header.Set("Accept-Encoding", "gzip")
		r.Header.Set("Accept", "text/plain")

		resp, err := http.DefaultClient.Do(r)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, resp.StatusCode)

		defer resp.Body.Close()

		zr, err := gzip.NewReader(resp.Body)
		require.NoError(t, err)

		b, err := io.ReadAll(zr)
		require.NoError(t, err)

		require.Equal(t, successBody, string(b))
	})
}
