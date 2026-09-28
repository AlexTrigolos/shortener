package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	mockW    http.ResponseWriter
	mockR    *http.Request
	mockURL  string
	mockCode int
)

func mockHTTPRedirect(w http.ResponseWriter, r *http.Request, url string, code int) {
	mockW = w
	mockR = r
	mockURL = url
	mockCode = code

	w.Header().Set("Location", url)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	body := "<a href=\"" + url + "\">Temporary Redirect</a>.\n"
	fmt.Fprintln(w, body)
}

func TestGetHandler(t *testing.T) {
	urls, err := model.NewURL()
	require.NoError(t, err)

	handler := chi.NewRouter()
	handler.Get("/{id}", GetHandler(urls))
	srv := httptest.NewServer(handler)

	urls.Set("short.url", srv.URL)

	defer srv.Close()

	httpRedirect = mockHTTPRedirect

	type want struct {
		statusCode  int
		contentType string
		response    string
		location    string
		redirect    bool
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name:    "нашлось значение по ключу",
			request: "/short.url",
			want: want{
				statusCode:  307,
				contentType: "text/html; charset=utf-8",
				response:    "",
				location:    srv.URL,
				redirect:    true,
			},
		},
		{
			name:    "нет записей с таким сокращением url",
			request: "/unknown.url",
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
			client := resty.New()
			client.SetRedirectPolicy(resty.NoRedirectPolicy()) // авторедирект делаем в ошибку
			req := client.R()
			req.Method = http.MethodGet
			req.URL = srv.URL + tt.request

			resp, err := req.Send()

			if tt.want.redirect {
				require.ErrorContains(t, err, "Get \""+tt.want.location+"\": auto redirect is disabled")
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
			assert.Equal(t, tt.want.contentType, resp.Header().Get("Content-Type"))
			assert.Equal(t, tt.want.response, string(resp.Body()))

			if tt.want.redirect {
				assert.Equal(t, tt.want.location, resp.Header().Get("Location"))
			} else {
				assert.Empty(t, resp.Header().Get("Location"))
			}
		})
	}
}
