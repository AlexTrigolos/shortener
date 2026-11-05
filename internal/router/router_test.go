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

type MockDB struct{}

func (mdb MockDB) PingContext(ctx context.Context) error {
	return nil
}

func TestNewGenerateRouter(t *testing.T) {
	t.Run("создать новый GenerateRouter", func(t *testing.T) {
		urls1, err := model.NewURL()
		require.NoError(t, err)
		newGenerateRouter := NewGenerateRouter(MockDB{}, urls1)

		urls2, err := model.NewURL()
		require.NoError(t, err)
		require.IsType(t, &GenerateRouter{newURL: urls2}, newGenerateRouter)
	})
}

func TestGenerateRoute(t *testing.T) {
	t.Run("удачный сгенерировали роуты", func(t *testing.T) {
		urls, err := model.NewURL()
		require.NoError(t, err)
		gr := NewGenerateRouter(MockDB{}, urls)
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
		req := resty.New().R()
		requests := [5][2]string{{http.MethodGet, ""}, {http.MethodPost, ""}, {http.MethodGet, "/ping"}, {http.MethodGet, "/short"}, {http.MethodPost, "/api/shorten"}}
		for _, request := range requests {
			req.Method = request[0]
			req.URL = srv.URL + request[1]

			resp, err := req.Send()
			require.NoError(t, err, "не удалось сделать http запрос")

			assert.NotEqual(t, http.StatusNotFound, resp.StatusCode(), "http страница не найдена")
		}
	})
}
