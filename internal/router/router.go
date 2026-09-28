package router

import (
	"github.com/AlexTrigolos/shortener/internal/handler"
	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

var newURL = model.NewURL

func GenerateRouter() (chi.Router, error) {
	r := chi.NewRouter()

	urls, err := newURL()
	if err != nil {
		return r, err
	}

	r.Use(middleware.ClientIPFromRemoteAddr, middleware.Logger, middleware.Recoverer)

	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.StatusHandler)
		r.Post("/", handler.CompactHandler(urls))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handler.GetHandler(urls))
		})
	})

	return r, nil
}
