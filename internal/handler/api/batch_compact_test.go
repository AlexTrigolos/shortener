package api

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
	batchShortURL = "AwHXdJfB"
	batchBaseURL  = "http://localhost:8080"
)

type MockBatchGenerator struct{}

func (mg MockBatchGenerator) BatchGenerateAndSetUniqURL(ctx context.Context, originalURLs *[]model.BatchRequest, baseURL string) (*[]model.BatchResponse, error) {
	resp := []model.BatchResponse{}
	for _, o := range *originalURLs {
		resp = append(resp, model.BatchResponse{CorrelationID: o.CorrelationID, ShortURL: batchBaseURL + "/" + batchShortURL})
	}
	return &resp, nil
}

type MockBatchFailGenerator struct{}

func (mg MockBatchFailGenerator) BatchGenerateAndSetUniqURL(ctx context.Context, originalURLs *[]model.BatchRequest, baseURL string) (*[]model.BatchResponse, error) {
	return &[]model.BatchResponse{}, errors.New("исчерпано количество попыток создать короткую ссылку")
}

func TestNewBatchCompact(t *testing.T) {
	logger := &zap.Logger{}

	urls, err := model.NewURL()
	require.NoError(t, err)
	generator := service.NewBatchGenerateURL(randstr.String, urls)
	t.Run("создать новый BatchCompact", func(t *testing.T) {
		newBatchCompact := NewBatchCompact(generator, logger)

		expectedCompact := &BatchCompact{generator: generator, logger: logger}
		require.IsType(t, expectedCompact, newBatchCompact)
		assert.Equal(t, reflect.ValueOf(expectedCompact.generator), reflect.ValueOf(newBatchCompact.generator))
		assert.Equal(t, expectedCompact.logger, newBatchCompact.logger)
	})
}

func TestBatchCompactHandler(t *testing.T) {
	handler := chi.NewRouter()

	type want struct {
		statusCode  int
		contentType string
		response    string
	}
	tests := []struct {
		name            string
		body            string
		generatorStruct BatchURLGenerator
		want            want
	}{
		{
			name:            "удачное сокращение",
			body:            "[{\"correlation_id\":\"1\",\"original_url\":\"http://long.url\"},{\"correlation_id\":\"2\",\"original_url\":\"http://long_sec.url\"}]",
			generatorStruct: MockBatchGenerator{},
			want: want{
				statusCode:  201,
				contentType: "application/json",
				response:    "[{\"correlation_id\":\"1\",\"short_url\":\"" + batchBaseURL + "/" + batchShortURL + "\"},{\"correlation_id\":\"2\",\"short_url\":\"" + batchBaseURL + "/" + batchShortURL + "\"}]\n",
			},
		},
		{
			name:            "передан не json",
			body:            "",
			generatorStruct: MockBatchGenerator{},
			want: want{
				statusCode:  400,
				contentType: "text/plain; charset=utf-8",
				response:    "не можем декодировать JSON тело запроса\n",
			},
		},
		{
			name:            "передан пустой original_url",
			body:            "[{\"correlation_id\":\"1\",\"original_url\":\"\"}]",
			generatorStruct: MockBatchGenerator{},
			want: want{
				statusCode:  400,
				contentType: "text/plain; charset=utf-8",
				response:    "переданный вами 1 по счету original_url пустой\n",
			},
		},
		{
			name:            "передан пустой correlation_id",
			body:            "[{\"correlation_id\":\"\",\"original_url\":\"http://long.url\"}]",
			generatorStruct: MockBatchGenerator{},
			want: want{
				statusCode:  400,
				contentType: "text/plain; charset=utf-8",
				response:    "переданный вами 1 по счету correlation_id пустой\n",
			},
		},
		{
			name:            "передан слишком длинный url",
			body:            "[{\"correlation_id\":\"\",\"original_url\":\"" + strings.Repeat("a", 1000) +"\"}]",
			generatorStruct: MockBatchGenerator{},
			want: want{
				statusCode:  413,
				contentType: "text/plain; charset=utf-8",
				response:    "тело запроса слишком большое\n",
			},
		},
		{
			name:            "сломался генератор",
			body:            "[{\"correlation_id\":\"1\",\"original_url\":\"http://long.url\"}]",
			generatorStruct: MockBatchFailGenerator{},
			want: want{
				statusCode:  500,
				contentType: "text/plain; charset=utf-8",
				response:    "Internal Server Error\n",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			apiCompact := NewBatchCompact(tt.generatorStruct, zap.NewNop())
			handler.Post("/api/shorten/batch", apiCompact.BatchCompactHandler(batchBaseURL))
			srv := httptest.NewServer(handler)
			defer srv.Close()

			req := resty.New().R()
			req.Method = http.MethodPost
			req.URL = srv.URL + "/api/shorten/batch"
			req.Body = tt.body

			resp, err := req.Send()
			require.NoError(t, err)

			assert.Equal(t, tt.want.statusCode, resp.StatusCode())
			assert.Equal(t, tt.want.contentType, resp.Header().Get("Content-Type"))

			assert.Equal(t, tt.want.response, string(resp.Body()))
		})
	}
}
