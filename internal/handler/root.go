package handler

import (
	"net/http"

	"github.com/AlexTrigolos/shortener/internal/model"
)

func Root(w http.ResponseWriter, r *http.Request, urls *model.URL) {
	switch r.Method {
	case http.MethodPost:
		Compact(w, r, urls)
	case http.MethodGet:
		Status(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
