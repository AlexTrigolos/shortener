package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	if err == nil {
		r.size += size
		if r.status == 0 {
			r.status = http.StatusOK
		}
	}
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.status = statusCode
}

func LoggerMiddleware(logger *zap.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lw := loggingResponseWriter{
				ResponseWriter: w,
				status:         0,
				size:           0,
			}

			h.ServeHTTP(&lw, r)

			duration := time.Since(start)

			logger.Info(
				"Data: ", zap.String("uri", r.RequestURI), zap.String("method", r.Method), zap.Duration("duration", duration),
				zap.Int("status", lw.status), zap.Int("size", lw.size),
			)
		})
	}
}
