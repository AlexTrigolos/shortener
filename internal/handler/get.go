package handler

import (
	"net/http"

	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/go-chi/chi/v5"
)

var httpRedirect = http.Redirect

func GetHandler(urls *model.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			http.Error(w, "не передан URL", http.StatusBadRequest)
			return
		}

		url, err := urls.Get(id)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		httpRedirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
