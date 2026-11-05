package handler

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockSuccessDB struct{}

func (mdb MockSuccessDB) PingContext(ctx context.Context) error {
	return nil
}

type MockFailDB struct{}

func (mdb MockFailDB) PingContext(ctx context.Context) error {
	return errors.New("ошибка")
}

func TestNewPing(t *testing.T) {
	logger := &zap.Logger{}
	t.Run("создать новый Ping", func(t *testing.T) {
		newPing := NewPing(MockSuccessDB{}, logger)

		expectedPing := &Ping{db: MockSuccessDB{}, logger: logger}
		require.IsType(t, expectedPing, newPing)
		assert.Equal(t, expectedPing.db, newPing.db)
		assert.Equal(t, expectedPing.logger, newPing.logger)
	})
}

func TestPingHandler(t *testing.T) {
	handler := chi.NewRouter()

	tests := []struct {
		name           string
		dataBase       DB
		wantStatusCode int
		wantBody       string
	}{
		{
			name:           "бд хорошая",
			dataBase:       MockSuccessDB{},
			wantStatusCode: 200,
			wantBody:       "DB still not dead :D",
		},
		{
			name:           "ошибка в бд",
			dataBase:       MockFailDB{},
			wantStatusCode: 500,
			wantBody:       "Internal Server Error\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ping := NewPing(tt.dataBase, zap.NewNop())
			handler.Get("/ping", ping.PingHandler())
			srv := httptest.NewServer(handler)
			defer srv.Close()

			req := resty.New().R()
			req.Method = http.MethodGet
			req.URL = srv.URL + "/ping"

			resp, err := req.Send()
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatusCode, resp.StatusCode())
			assert.Equal(t, tt.wantBody, string(resp.Body()))
		})
	}
}
