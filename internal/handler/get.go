package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

type Getter interface {
	Get(ctx context.Context, key string) (string, error)
}

func GetHandler(urls Getter, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		url, err := urls.Get(ctx, id)
		if err != nil && errors.Is(err, model.ErrDeleted) {
			logger.Error("ссылка удалена создателем", zap.String("Error", err.Error()))
			http.Error(w, http.StatusText(http.StatusGone), http.StatusGone)
			return
		}

		if err != nil {
			http.Error(w, "Не нашли запись для: "+id, http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
