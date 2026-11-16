package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

type Getter interface {
	Get(ctx context.Context, key string) (string, error)
}

func GetHandler(urls Getter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		url, err := urls.Get(ctx, id)
		if err != nil {
			http.Error(w, "Не нашли запись для: "+id, http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
