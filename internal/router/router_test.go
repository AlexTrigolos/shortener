package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/config"
	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockedDBURL struct{}

func (mdb MockedDBURL) Set(ctx context.Context, key, value string) error {
	return nil
}
func (mdb MockedDBURL) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}
func (mdb MockedDBURL) GetByOriginal(ctx context.Context, original string) (string, error) {
	return "", nil
}
func (mdb MockedDBURL) PingContext(ctx context.Context) error {
	return nil
}
func (mdb MockedDBURL) Begin(int) (*model.Transaction, error) {
	return nil, nil
}
func (mdb MockedDBURL) SetTx(ctx context.Context, key string, value string) error {
	return nil
}
func (mdb MockedDBURL) Close() error {
	return nil
}

func TestNewGenerateRouter(t *testing.T) {
	t.Run("создать новый GenerateRouter", func(t *testing.T) {
		urls1, err := model.NewURL()
		require.NoError(t, err)
		newGenerateRouter := NewGenerateRouter(urls1)

		urls2, err := model.NewURL()
		require.NoError(t, err)
		require.IsType(t, &GenerateRouter{URLs: urls2}, newGenerateRouter)
	})
}

func TestGenerateRoute(t *testing.T) {
	t.Run("удачный сгенерировали роуты", func(t *testing.T) {
		gr := NewGenerateRouter(&MockedDBURL{})
		server, err := config.PrepareServer()
		require.NoError(t, err)
		router, err := gr.GenerateRoute("flagbaseURL", server)

		require.NoError(t, err)

		srv := httptest.NewServer(router)
		defer srv.Close()

		// Get("/", handler.StatusHandler)
		// Post("/", handler.CompactHandler(baseURL)
		// Get("/ping", handler.PingHandler)
		// Get("/{id}", handler.GetHandler(newURL))
		// Post("/api/shorten", handler.CompactHandler(baseURL)
		// Post("/api/shorten/batch", handler.BatchCompactHandler(baseURL)
		req := resty.New().R()
		requests := [6][2]string{
			{http.MethodGet, ""},
			{http.MethodPost, ""},
			{http.MethodGet, "/ping"},
			{http.MethodGet, "/short"},
			{http.MethodPost, "/api/shorten"},
			{http.MethodPost, "/api/shorten/batch"},
		}
		for _, request := range requests {
			req.Method = request[0]
			req.URL = srv.URL + request[1]

			resp, err := req.Send()
			require.NoError(t, err)

			assert.NotEqual(t, http.StatusNotFound, resp.StatusCode(), "http страница не найдена")
			assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode(), "ошибка выполнения")
		}
	})
}
