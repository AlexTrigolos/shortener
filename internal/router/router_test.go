package router

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func mockSuccessNewURL() (*model.URL, error) {
	urls, _ := model.NewURL()
	urls.Set("short", "http://google.com") // для удачной проверки Get урла
	return urls, nil
}

func mockFailNewURL() (*model.URL, error) {
	urls, _ := model.NewURL()
	return urls, errors.New("не создали urls")
}

func TestRoute(t *testing.T) {
	tests := []struct {
		name    string
		mockURL func() (*model.URL, error)
		err     string
	}{
		{
			name:    "удачный сгенерировали роуты",
			mockURL: mockSuccessNewURL,
			err:     "",
		},
		{
			name:    "не смогли создать urls",
			mockURL: mockFailNewURL,
			err:     "не создали urls",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newURL = tt.mockURL
			router, err := GenerateRouter()

			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
			} else {
				require.NoError(t, err)

				srv := httptest.NewServer(router)
				defer srv.Close()

				// Get("/", handler.StatusHandler)
				// Post("/", handler.CompactHandler(urls)
				// Get("/{id}", handler.GetHandler(urls))
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
