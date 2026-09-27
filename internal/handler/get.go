package handler

import (
	"net/http"
	"strings"

	"github.com/AlexTrigolos/shortener/internal/model"
)

func GetHandler(urls *model.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "ожидался Get запрос", http.StatusMethodNotAllowed)
			return
		}

		id := strings.Trim(r.URL.Path, "/")
		if id == "" {
			http.Error(w, "не передан URL", http.StatusBadRequest)
			return
		}

		url, err := urls.Get(id)
		if err != nil {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Location", url)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
