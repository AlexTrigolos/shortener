package router

import (
	"github.com/AlexTrigolos/shortener/internal/handler"
	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/AlexTrigolos/shortener/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/thanhpk/randstr"
)

type NewURL model.GetterSetter

type GenerateRouter struct {
	newURL NewURL
}

func NewGenerateRouter(nu model.GetterSetter) *GenerateRouter {
	return &GenerateRouter{newURL: nu}
}

func (gr *GenerateRouter) GenerateRoute(baseURL string) (chi.Router, error) {
	r := chi.NewRouter()

	r.Use(middleware.ClientIPFromRemoteAddr, middleware.Logger, middleware.Recoverer)

	generateURL := service.NewGenerateURL(randstr.String)
	compact := handler.NewCompact(generateURL.GenerateAndSetUniqURL)

	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.StatusHandler)
		r.Post("/", compact.CompactHandler(gr.newURL, baseURL))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handler.GetHandler(gr.newURL))
		})
	})

	return r, nil
}
