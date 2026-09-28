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

const shortURL = "EwHXdJfB"

func mockGenerateUniqURL(urls *model.URL) string {
	return shortURL
}

func TestCompactHandler(t *testing.T) {
	urls, err := model.NewURL()
	require.NoError(t, err)

	handler := chi.NewRouter()
	handler.Post("/", CompactHandler(urls))
	srv := httptest.NewServer(handler)

	urls.Set("short.url", srv.URL)
	generateUniqURL = mockGenerateUniqURL

	defer srv.Close()

	type want struct {
		statusCode  int
		contentType string
		response    string
	}
	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "удачное сокращение",
			body: "http://long.url",
			want: want{
				statusCode:  201,
				contentType: "text/plain",
				response:    "http://localhost:8080/" + shortURL,
			},
		},
		{
			name: "передан пустой url",
			body: "",
			want: want{
				statusCode:  400,
				contentType: "text/plain; charset=utf-8",
				response:    "переданный вами URL пустой\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := resty.New().R()
			req.Method = http.MethodPost
			req.URL = srv.URL
			req.Body = tt.body

			resp, err := req.Send()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
			assert.Equal(t, tt.want.contentType, resp.Header().Get("Content-Type"))

			assert.Equal(t, tt.want.response, string(resp.Body()))
		})
	}
}
