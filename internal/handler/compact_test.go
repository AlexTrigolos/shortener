package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompactHandler(t *testing.T) {
	type want struct {
		statusCode  int
		contentType string
		response    string
	}
	tests := []struct {
		name     string
		keyValue [2]string
		method   string
		body     string
		want     want
	}{
		{
			name:     "удачное сокращение",
			keyValue: [2]string{"short.url", "http://another.url"},
			method:   http.MethodPost,
			body:     "http://long.url",
			want: want{
				statusCode:  201,
				contentType: "text/plain",
				response:    "http://localhost:8080/EwHXdJfB",
			},
		},
		{
			name:     "не тот тип запроса",
			keyValue: [2]string{"short.url", "http://another.url"},
			method:   http.MethodPut,
			body:     "http://long.url",
			want: want{
				statusCode:  405,
				contentType: "text/plain; charset=utf-8",
				response:    "ожидался Post запрос\n",
			},
		},
		{
			name:     "передан пустой url",
			keyValue: [2]string{"short.url", "http://another.url"},
			method:   http.MethodPost,
			body:     "",
			want: want{
				statusCode:  400,
				contentType: "text/plain; charset=utf-8",
				response:    "переданный вами URL пустой\n",
			},
		},
		{
			name:     "короткий путь уже занят",
			keyValue: [2]string{"EwHXdJfB", "http://another.url"},
			method:   http.MethodPost,
			body:     "http://long.url",
			want: want{
				statusCode:  500,
				contentType: "text/plain; charset=utf-8",
				response:    "ошибка добавления URL\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			urls, err := model.NewURL()
			require.NoError(t, err)
			urls.Set(tt.keyValue[0], tt.keyValue[1])

			r := httptest.NewRequest(tt.method, "/", bytes.NewBuffer([]byte(tt.body)))
			w := httptest.NewRecorder()
			CompactHandler(w, r, urls)

			res := w.Result()
			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

			assert.Equal(t, tt.want.response, w.Body.String())
		})
	}
}
