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
	mockW          http.ResponseWriter
	mockR          *http.Request
	mockURL        string
	mockCode       int
	flagDefaultURL = "http://flagDefaultURL.ru"
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
	handler.Get("/{id}", GetHandler(urls, flagDefaultURL))
	srv := httptest.NewServer(handler)

	urls.Set("short.url", srv.URL)

	defer srv.Close()

	httpRedirect = mockHTTPRedirect

	tests := []struct {
		name          string
		request       string
		wantLocation string
	}{
		{
			name:          "нашлось значение по ключу",
			request:       "/short.url",
			wantLocation: srv.URL,
		},
		{
			name:          "нет записей с таким сокращением url",
			request:       "/unknown.url",
			wantLocation: flagDefaultURL,
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

			require.ErrorContains(t, err, "Get \""+tt.wantLocation+"\": auto redirect is disabled")

			assert.Equal(t, 307, resp.StatusCode())
			assert.Equal(t, "text/html; charset=utf-8", resp.Header().Get("Content-Type"))
			assert.Equal(t, "", string(resp.Body()))

			assert.Equal(t, tt.wantLocation, resp.Header().Get("Location"))
		})
	}
}
