package app

import (
	"net/http"

	"github.com/AlexTrigolos/shortener/internal/handler"
	"github.com/AlexTrigolos/shortener/internal/model"
)

func Run() error {
	urls := model.NewURL()
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, func(w http.ResponseWriter, r *http.Request) {
		handler.Root(w, r, urls)
	})
	mux.HandleFunc(`/{id}`, func(w http.ResponseWriter, r *http.Request) {
		handler.Get(w, r, urls)
	})

	return http.ListenAndServe(`:8080`, mux)
}
