package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/AlexTrigolos/shortener/internal/model"
	"go.uber.org/zap"
)

type URLGenerator interface {
	GenerateAndSetUniqURL(ctx context.Context, originalURL string) (string, error)
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
		var req model.Request
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
		defer r.Body.Close()
		if err := dec.Decode(&req); err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				http.Error(w, "тело запроса слишком большое", http.StatusRequestEntityTooLarge)
				return
			}

			_, err := io.ReadAll(r.Body)

			if err != nil {
				http.Error(w, "не удалось прочитать тело запроса", http.StatusBadRequest)
				return
			}

			http.Error(w, "не можем декодировать JSON тело запроса", http.StatusBadRequest)
			return
		}

		originalURL := string(req.URL)
		if originalURL == "" {
			http.Error(w, "переданный вами URL пустой", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		short, err := c.generator.GenerateAndSetUniqURL(ctx, originalURL)
		if err != nil {
			c.logger.Error("создание короткой ссылки", zap.String("Error", err.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		resp := model.Response{Result: baseURL + "/" + short}
		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			c.logger.Error("сжатие ответа", zap.String("Error", err.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}
