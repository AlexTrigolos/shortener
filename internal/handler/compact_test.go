package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/AlexTrigolos/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thanhpk/randstr"
	"go.uber.org/zap"
)

const (
	shortURL = "EwHXdJfB"
	baseURL  = "http://localhost:8080"
)

type MockGenerator struct{}

func (mg MockGenerator) GenerateAndSetUniqURL(context.Context, string) (string, error) {
	return shortURL, nil
}

type MockFailGenerator struct{}

func (mg MockFailGenerator) GenerateAndSetUniqURL(context.Context, string) (string, error) {
	return "", errors.New("исчерпано количество попыток создать короткую ссылку")
}

func TestNewCompact(t *testing.T) {
	logger := &zap.Logger{}

	urls, err := model.NewURL()
	require.NoError(t, err)
	generator := service.NewGenerateURL(randstr.String, urls)
	t.Run("создать новый Compact", func(t *testing.T) {
		newCompact := NewCompact(generator, logger)

		expectedCompact := &Compact{generator: generator, logger: logger}
		require.IsType(t, expectedCompact, newCompact)
		assert.Equal(t, reflect.ValueOf(expectedCompact.generator), reflect.ValueOf(newCompact.generator))
		assert.Equal(t, expectedCompact.logger, newCompact.logger)
	})
}

func TestCompactHandler(t *testing.T) {
	urls, err := model.NewURL()
	require.NoError(t, err)

	handler := chi.NewRouter()

	type want struct {
		statusCode  int
		contentType string
		response    string
	}
	tests := []struct {
		name            string
		body            string
		generatorStruct URLGenerator
		want            want
	}{
		{
			name:            "удачное сокращение",
			body:            "http://long.url",
			generatorStruct: MockGenerator{},
			want: want{
				statusCode:  201,
				contentType: "text/plain",
				response:    baseURL + "/" + shortURL,
			},
		},
		{
			name:            "передан пустой url",
			body:            "",
			generatorStruct: MockGenerator{},
			want: want{
				statusCode:  400,
				contentType: "text/plain; charset=utf-8",
				response:    "переданный вами URL пустой\n",
			},
		},
		{
			name:            "передан слишком длинный url",
			body:            strings.Repeat("a", 1025),
			generatorStruct: MockGenerator{},
			want: want{
				statusCode:  413,
				contentType: "text/plain; charset=utf-8",
				response:    "тело запроса слишком большое\n",
			},
		},
		{
			name:            "сломался генератор",
			body:            "http://long.url",
			generatorStruct: MockFailGenerator{},
			want: want{
				statusCode:  500,
				contentType: "text/plain; charset=utf-8",
				response:    "Internal Server Error\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiCompact := NewCompact(tt.generatorStruct, zap.NewNop())
			handler.Post("/", apiCompact.CompactHandler(baseURL))
			srv := httptest.NewServer(handler)
			defer srv.Close()

			urls.Set(t.Context(), "short.url", srv.URL)

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
