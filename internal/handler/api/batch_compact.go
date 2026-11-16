package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/AlexTrigolos/shortener/internal/model"
	"go.uber.org/zap"
)

type BatchURLGenerator interface {
	BatchGenerateAndSetUniqURL(ctx context.Context, originalURLs *[]model.BatchRequest, baseURL string) (*[]model.BatchResponse, error)
}

type BatchCompact struct {
	generator BatchURLGenerator
	logger    *zap.Logger
}

func NewBatchCompact(ug BatchURLGenerator, logger *zap.Logger) *BatchCompact {
	return &BatchCompact{generator: ug, logger: logger}
}

func (c BatchCompact) BatchCompactHandler(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req []model.BatchRequest
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

		for i, r := range req {
			URL := string(r.OriginalURL)
			if URL == "" {
				errStr := "переданный вами " + strconv.Itoa(i + 1) + " по счету original_url пустой"
				http.Error(w, errStr, http.StatusBadRequest)
				return
			}
			corr := string(r.CorrelationID)
			if corr == "" {
				errStr := "переданный вами " + strconv.Itoa(i + 1) + " по счету correlation_id пустой"
				http.Error(w, errStr, http.StatusBadRequest)
				return
			}
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		resp, err := c.generator.BatchGenerateAndSetUniqURL(ctx, &req, baseURL)
		if err != nil {
			c.logger.Error("создание короткой ссылки", zap.String("Error", err.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			c.logger.Error("сжатие ответа", zap.String("Error", err.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}
