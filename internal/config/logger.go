package config

import (
	"fmt"
	"net/http"
	"strconv"
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

func WithLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger, err := zap.NewDevelopment()
		if err != nil {
			panic(err)
		}
		defer logger.Sync()

		sugar := *logger.Sugar()

		start := time.Now()

		lw := loggingResponseWriter{
			ResponseWriter: w,
			status: 0,
			size:   0,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		sugar.Infoln(
			"Request Data: uri -", r.RequestURI+",", "method -", r.Method+",", "duration -", fmt.Sprint(duration)+".",
			"Response Data: status -", strconv.Itoa(lw.status)+",", "size -", strconv.Itoa(lw.size)+".",
		)
	})
}
