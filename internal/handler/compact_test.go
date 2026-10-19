package handler

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/AlexTrigolos/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thanhpk/randstr"
)

const (
	shortURL = "EwHXdJfB"
	baseURL  = "http://localhost:8080"
)

type MockGenerator struct {}

func (mg MockGenerator) GenerateAndSetUniqURL(string) (string, error) {
	return shortURL, nil
}

func TestNewCompact(t *testing.T) {
	urls, err := model.NewURL()
	require.NoError(t, err)
	generator := service.NewGenerateURL(randstr.String, urls)
	t.Run("создать новый Compact", func(t *testing.T) {
		newCompact := NewCompact(generator)

		expectedCompact := &Compact{generator: generator, logger: slog.New(slog.NewTextHandler(os.Stdout, nil))}
		require.IsType(t, expectedCompact, newCompact)
		assert.Equal(t, reflect.ValueOf(expectedCompact.generator), reflect.ValueOf(newCompact.generator))
		assert.Equal(t, expectedCompact.logger, newCompact.logger)
	})
}

func TestCompactHandler(t *testing.T) {
	urls, err := model.NewURL()
	require.NoError(t, err)

	handler := chi.NewRouter()
	compact := NewCompact(MockGenerator{})
	handler.Post("/", compact.CompactHandler(baseURL))
	srv := httptest.NewServer(handler)

	urls.Set("short.url", srv.URL)

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
				response:    baseURL + "/" + shortURL,
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
