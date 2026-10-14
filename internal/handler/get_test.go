package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHandler(t *testing.T) {
	urls, err := model.NewURL()
	require.NoError(t, err)

	handler := chi.NewRouter()
	handler.Get("/{id}", GetHandler(urls))
	srv := httptest.NewServer(handler)

	urls.Set("short.url", srv.URL)

	defer srv.Close()

	type Want struct {
		location    string
		statusCode  int
		contentType string
		response    string
	}

	tests := []struct {
		name    string
		request string
		want    Want
	}{
		{
			name:    "нашлось значение по ключу",
			request: "/short.url",
			want: Want{
				location:    srv.URL,
				statusCode:  307,
				contentType: "text/html; charset=utf-8",
				response:    "",
			},
		},
		{
			name:    "нет записей с таким сокращением url",
			request: "/unknown.url",
			want: Want{
				location:    "",
				statusCode:  400,
				contentType: "text/plain; charset=utf-8",
				response:    "Не нашли запись для: unknown.url\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			client.SetRedirectPolicy(resty.NoRedirectPolicy()) // авторедирект делаем в ошибку
			req := client.R()
			req.Method = http.MethodGet
			req.URL = srv.URL + tt.request

			resp, err := req.Send()

			if tt.want.location == "" {
				require.NoError(t, err)

				assert.Equal(t, tt.want.statusCode, resp.StatusCode())
				assert.Equal(t, tt.want.contentType, resp.Header().Get("Content-Type"))

				assert.Equal(t, tt.want.response, string(resp.Body()))
			} else {
				require.ErrorContains(t, err, "Get \""+tt.want.location+"\": auto redirect is disabled")

				assert.Equal(t, tt.want.statusCode, resp.StatusCode())
				assert.Equal(t, tt.want.contentType, resp.Header().Get("Content-Type"))
				assert.Equal(t, tt.want.response, string(resp.Body()))

				assert.Equal(t, tt.want.location, resp.Header().Get("Location"))
			}
		})
	}
}
