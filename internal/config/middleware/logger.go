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
	r.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.status = statusCode
}

func LoggerMiddleware(logger *zap.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sugar := *logger.Sugar()

			start := time.Now()

			lw := loggingResponseWriter{
				ResponseWriter: w,
				status:         0,
				size:           0,
			}

			h.ServeHTTP(&lw, r)

			duration := time.Since(start)

			sugar.Infoln(
				"Request Data:", zap.String("uri", r.RequestURI), zap.String("method", r.Method), zap.Duration("duration", duration),
				"Response Data:", zap.Int("status", lw.status), zap.Int("size", lw.size),
			)
		})
	}
}
