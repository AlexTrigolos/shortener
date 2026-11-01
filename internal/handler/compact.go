package handler

import (
	"io"
	"net/http"

	"go.uber.org/zap"
)

type URLGenerator interface {
	GenerateAndSetUniqURL(originalURL string) (string, error)
}

type Compact struct {
	generator URLGenerator
	logger    *zap.Logger
}

func NewCompact(ug URLGenerator, logger *zap.Logger) *Compact {
	return &Compact{generator: ug, logger: logger}
}

func (c Compact) CompactHandler(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1024))
		if err != nil {
			if err.Error() == "http: request body too large" {
				http.Error(w, "тело запроса слишком большое", http.StatusRequestEntityTooLarge)
				return
			}

			http.Error(w, "не удалось прочитать тело запроса", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		originalURL := string(body)
		if originalURL == "" {
			http.Error(w, "переданный вами URL пустой", http.StatusBadRequest)
			return
		}

		short, err := c.generator.GenerateAndSetUniqURL(originalURL)
		if err != nil {
			c.logger.Error("создание короткой ссылки", zap.String("Error", err.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(baseURL + "/" + short))
	}
}
