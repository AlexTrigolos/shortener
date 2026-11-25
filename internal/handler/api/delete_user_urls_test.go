package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/AlexTrigolos/shortener/internal/auth"
	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDeleteUserURLsHandler(t *testing.T) {
	type Want struct {
		statusCode  int
		contentType string
		response    string
	}

	tests := []struct {
		name   string
		body   string
		userID string
		want   Want
	}{
		{
			name:   "смогли удалить записи",
			body:   "[\"short_1\", \"short_2\"]",
			userID: "user_id",
			want: Want{
				statusCode:  http.StatusAccepted,
				contentType: "text/plain; charset=utf-8",
				response:    "Accepted\n",
			},
		},
		{
			name:   "нет доступа",
			body:   "[\"short_1\", \"short_2\"]",
			userID: "unknown_user_id",
			want: Want{
				statusCode:  http.StatusBadRequest,
				contentType: "text/plain; charset=utf-8",
				response:    "Bad Request\n",
			},
		},
		{
			name:   "передан не json",
			body:   "",
			userID: "user_id",
			want: Want{
				statusCode:  400,
				contentType: "text/plain; charset=utf-8",
				response:    "не можем декодировать JSON тело запроса\n",
			},
		},
		{
			name: "передан слишком длинный url",
			body: "[\"" + strings.Repeat("a", 1024) + "\"]",
			want: Want{
				statusCode:  413,
				contentType: "text/plain; charset=utf-8",
				response:    "тело запроса слишком большое\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			urls, err := model.NewURL()
			require.NoError(t, err)
			doneCh := make(chan struct{})
			deleter, _ := model.NewDeleter(doneCh, urls, zap.NewNop())
			t.Cleanup(func() { close(doneCh) })

			handler := chi.NewRouter()

			err = urls.Set(t.Context(), "short_1", "https://google.com", "user_id")
			require.NoError(t, err)

			err = urls.Set(t.Context(), "short_2", "https://yandex.ru", "user_id")
			require.NoError(t, err)

			err = urls.Set(t.Context(), "short_3", "https://yandex-go.com", "user_id")
			require.NoError(t, err)

			err = urls.Set(t.Context(), "short_4", "https://yandex-practicum.ru", "user_id_2")
			require.NoError(t, err)

			handler.Delete("/api/user/urls", func(w http.ResponseWriter, r *http.Request) {
				DeleteUserURLsHandler(urls, deleter, zap.NewNop()).ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), auth.CookieName, tt.userID)))
			})
			srv := httptest.NewServer(handler)
			defer srv.Close()

			req := resty.New().R()
			req.Method = http.MethodDelete
			req.URL = srv.URL + "/api/user/urls"
			req.Body = tt.body

			resp, err := req.Send()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
			assert.Equal(t, tt.want.contentType, resp.Header().Get("Content-Type"))

			assert.Equal(t, tt.want.response, string(resp.Body()))

			// ждем выполнения
			time.Sleep(11 * time.Second)
			if tt.want.statusCode == http.StatusAccepted {
				_, err := urls.Get(t.Context(), "short_1")
				require.Error(t, err)
			} else {
				originalURL, err := urls.Get(t.Context(), "short_1")
				require.NoError(t, err)
				assert.Contains(t, "https://google.com", originalURL)
			}
		})
	}
}
