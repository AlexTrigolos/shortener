package handler

import (
	"net/http"

	"github.com/AlexTrigolos/shortener/internal/model"
)

func Get(w http.ResponseWriter, r *http.Request, urls *model.URL) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	
	id := r.PathValue("id")

	url := urls.Get(id)

	w.Header().Set("Location", url)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
