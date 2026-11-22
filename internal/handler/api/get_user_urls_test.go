package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/auth"
	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const userBaseURL = "http://localhost:8080"

func TestGetUserURLsHandler(t *testing.T) {
	urls, err := model.NewURL()
	require.NoError(t, err)

	handler := chi.NewRouter()

	err = urls.Set(t.Context(), "short.url", "https://google.com", "user_id")
	require.NoError(t, err)

	type Want struct {
		statusCode  int
		contentType string
		response    string
	}

	tests := []struct {
		name    string
		userID string
		want    Want
	}{
		{
			name:    "нашлось значение по ключу",
			userID: "user_id",
			want: Want{
				statusCode:  200,
				contentType: "application/json",
				response:    "[{\"short_url\":\"http://localhost:8080/short.url\",\"original_url\":\"https://google.com\"}]\n",
			},
		},
		{
			name:    "нет записей с таким сокращением url",
			userID: "unknown_user_id",
			want: Want{
				statusCode:  204,
				contentType: "text/plain; charset=utf-8",
				response:    "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler.Get("/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
				GetUserURLsHandler(userBaseURL, urls, zap.NewNop()).ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), auth.CookieName, tt.userID)))
			})
			srv := httptest.NewServer(handler)
			defer srv.Close()

			req := resty.New().R()
			req.Method = http.MethodGet
			req.URL = srv.URL + "/api/user/urls"

			resp, err := req.Send()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
			assert.Equal(t, tt.want.contentType, resp.Header().Get("Content-Type"))

			assert.Equal(t, tt.want.response, string(resp.Body()))
		})
	}
}
