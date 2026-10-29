package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_loggingResponseWriter_Write(t *testing.T) {
	tests := []struct {
		name string
		b    []byte
		want int
	}{
		{
			name: "нормальная строка",
			b:    []byte("abc"),
			want: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := loggingResponseWriter{ResponseWriter: httptest.NewRecorder()}
			_, gotErr := r.Write(tt.b)
			require.NoError(t, gotErr)
			assert.Equal(t, tt.want, r.size)
		})
	}
}

func Test_loggingResponseWriter_WriteHeader(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{
			name:       "удачный код",
			statusCode: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := loggingResponseWriter{ResponseWriter: httptest.NewRecorder()}
			r.WriteHeader(tt.statusCode)
			assert.Equal(t, tt.statusCode, r.status)
		})
	}
}

func fakeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`Server started :)`))
}

func TestLoggerMiddleware(t *testing.T) {
	tests := []struct {
		name        string
		handlerFunc func(w http.ResponseWriter, r *http.Request)
	}{
		{
			name:        "удачный лог",
			handlerFunc: fakeHandler,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(tt.handlerFunc)
			srv := httptest.NewServer(LoggerMiddleware(handler))
			defer srv.Close()

			req := resty.New().R()
			req.Method = http.MethodGet
			req.URL = srv.URL

			_, err := req.Send()
			require.NoError(t, err)
		})
	}
}
