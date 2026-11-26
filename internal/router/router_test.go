package router

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/config"
	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockedDBURL struct{}

func (mdb MockedDBURL) Set(ctx context.Context, key, value, userID string) error {
	return nil
}
func (mdb MockedDBURL) Get(ctx context.Context, key string) (string, error) {
	return "", nil
}
func (mdb MockedDBURL) GetByOriginal(ctx context.Context, original, baseURL string) (string, error) {
	return "", nil
}
func (mdb MockedDBURL) GetByUserID(_ context.Context, userID, baseURL string) (*[]model.UserURL, error) {
	return &[]model.UserURL{}, nil
}
func (mdb MockedDBURL) CheckUserIDURLs(ctx context.Context, urls []string, userID string) error {
	return nil
}
func (mdb MockedDBURL) DeleteURLs(ctx context.Context, urls []model.URLsData) error {
	return nil
}
func (mdb MockedDBURL) PingContext(ctx context.Context) error {
	return nil
}
func (mdb MockedDBURL) Begin(int) (*model.Transaction, error) {
	return nil, nil
}

func TestNewGenerateRouter(t *testing.T) {
	t.Run("создать новый GenerateRouter", func(t *testing.T) {
		urls1, err := model.NewURL()
		require.NoError(t, err)
		doneCh := make(chan struct{})
		deleter, _ := model.NewDeleter(doneCh, urls1, zap.NewNop())
		defer close(doneCh)
		newGenerateRouter := NewGenerateRouter(urls1, deleter)

		urls2, err := model.NewURL()
		require.NoError(t, err)
		require.IsType(t, &GenerateRouter{URLs: urls2}, newGenerateRouter)
	})
}

func TestGenerateRoute(t *testing.T) {
	os.Setenv("SECRET_KEY", "key")
	defer func() { os.Unsetenv("SECRET_KEY") }()

	t.Run("удачный сгенерировали роуты", func(t *testing.T) {
		doneCh := make(chan struct{})
		deleter, _ := model.NewDeleter(nil, &MockedDBURL{}, zap.NewNop())
		defer close(doneCh)
		gr := NewGenerateRouter(&MockedDBURL{}, deleter)
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
		// Post("/api/shorten", api_handler.CompactHandler(baseURL)
		// Post("/api/shorten/batch", api_handler.BatchCompactHandler(baseURL)
		// Get("/api/user/urls", api_handler.GetUserURLsHandler(baseURL, gr.URLs, server.Logger))
		// Delete("/api/user/urls", api_handler.GetUserURLsHandler(gr.URLs, gr.deleter, server.Logger))
		req := resty.New().R()
		requests := [8][2]string{
			{http.MethodGet, ""},
			{http.MethodPost, ""},
			{http.MethodGet, "/ping"},
			{http.MethodGet, "/short"},
			{http.MethodPost, "/api/shorten"},
			{http.MethodPost, "/api/shorten/batch"},
			{http.MethodGet, "/api/user/urls"},
			{http.MethodDelete, "/api/user/urls"},
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
