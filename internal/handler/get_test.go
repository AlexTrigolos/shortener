package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		response    string
		location    string
		redirect    bool
	}
	tests := []struct {
		name     string
		keyValue [2]string
		method   string
		request  string
		want     want
	}{
		{
			name:     "нашлось значение по ключу",
			keyValue: [2]string{"short.url", "https://long.url"},
			method:   http.MethodGet,
			request:  "/short.url",
			want: want{
				statusCode:  307,
				contentType: "text/html; charset=utf-8",
				response:    "<a href=\"https://long.url\">Temporary Redirect</a>.\n\n",
				location:    "https://long.url",
				redirect:    true,
			},
		},
		{
			name:     "вызван не тот метод",
			keyValue: [2]string{"short.url", "https://long.url"},
			method:   http.MethodPost,
			request:  "/short.url",
			want: want{
				statusCode:  405,
				contentType: "text/plain; charset=utf-8",
				response:    "ожидался Get запрос\n",
				location:    "",
				redirect:    false,
			},
		},
		{
			name:     "не передан ключ",
			keyValue: [2]string{"short.url", "https://long.url"},
			method:   http.MethodGet,
			request:  "/",
			want: want{
				statusCode:  400,
				contentType: "text/plain; charset=utf-8",
				response:    "не передан URL\n",
				location:    "",
				redirect:    false,
			},
		},
		{
			name:     "нет записей с таким сокращением url",
			keyValue: [2]string{"short.url", "https://long.url"},
			method:   http.MethodGet,
			request:  "/unknown.url",
			want: want{
				statusCode:  404,
				contentType: "text/plain; charset=utf-8",
				response:    "404 page not found\n",
				location:    "",
				redirect:    false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls, err := model.NewURL()
			require.NoError(t, err)
			urls.Set(tt.keyValue[0], tt.keyValue[1])

			r := httptest.NewRequest(tt.method, tt.request, nil)
			w := httptest.NewRecorder()
			GetHandler(urls)(w, r)

			res := w.Result()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

			assert.Equal(t, tt.want.response, w.Body.String())

			if tt.want.redirect {
				assert.Equal(t, tt.want.location, res.Header.Get("Location"))
			} else {
				assert.Empty(t, res.Header.Get("Location"))
			}
		})
	}
}
