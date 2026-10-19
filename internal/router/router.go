package router

import (
	"github.com/AlexTrigolos/shortener/internal/handler"
	"github.com/AlexTrigolos/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/thanhpk/randstr"
)

type GetterSetter interface {
	Set(key, value string) error
	Get(key string) (string, error)
}

type GenerateRouter struct {
	newURL GetterSetter
}

func NewGenerateRouter(nu GetterSetter) *GenerateRouter {
	return &GenerateRouter{newURL: nu}
}

func (gr *GenerateRouter) GenerateRoute(baseURL string) (chi.Router, error) {
	r := chi.NewRouter()

	r.Use(middleware.ClientIPFromRemoteAddr, middleware.Logger, middleware.Recoverer)

	generateURL := service.NewGenerateURL(randstr.String, gr.newURL)
	compact := handler.NewCompact(generateURL)

	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.StatusHandler)
		r.Post("/", compact.CompactHandler(baseURL))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handler.GetHandler(gr.newURL))
		})
	})

	return r, nil
}
