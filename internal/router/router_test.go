package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/interfaces"
	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockNewURL() func() (interfaces.GetterSetter, error) {
	return model.NewURL
}

func TestNewGenerateRouter(t *testing.T) {
	t.Run("создать новый GenerateRouter", func(t *testing.T) {
		newGenerateRouter := NewGenerateRouter(mockNewURL())

		require.IsType(t, &GenerateRouter{newURL: model.NewURL}, newGenerateRouter)
	})
}

func TestGenerateRoute(t *testing.T) {
	tests := []struct {
		name    string
		mockURL func() (interfaces.GetterSetter, error)
		err     string
	}{
		{
			name:    "удачный сгенерировали роуты",
			mockURL: mockNewURL(),
			err:     "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gr := NewGenerateRouter(tt.mockURL)
			router, err := gr.GenerateRoute("flagDefaultURL")

			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
			} else {
				require.NoError(t, err)

				srv := httptest.NewServer(router)
				defer srv.Close()

				// Get("/", handler.StatusHandler)
				// Post("/", handler.CompactHandler(urls)
				// Get("/{id}", handler.GetHandler(urls, flagDefaultURL))
				req := resty.New().R()
				requests := [3][2]string{{http.MethodGet, ""}, {http.MethodPost, ""}, {http.MethodGet, "/short"}}
				for _, request := range requests {
					req.Method = request[0]
					req.URL = srv.URL + request[1]

					resp, err := req.Send()
					require.NoError(t, err, "не удалось сделать http запрос")

					assert.NotEqual(t, http.StatusNotFound, resp.StatusCode(), "http страница не найдена")
				}
			}
		})
	}
}
