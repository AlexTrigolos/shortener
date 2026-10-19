package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Getter interface {
	Get(key string) (string, error)
}

func GetHandler(urls Getter) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		url, err := urls.Get(id)
		if err != nil {
			http.Error(w, "Не нашли запись для: "+id, http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
