package handler

import (
	"net/http"

	"github.com/AlexTrigolos/shortener/internal/model"
)

// необходимо было вынести в переменные, чтобы можно было замокать
var CompactHndlr = CompactHandler
var StatusHndlr = StatusHandler

func RootHandler(urls *model.URL) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			CompactHndlr(w, r, urls)
		case http.MethodGet:
			StatusHndlr(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}
