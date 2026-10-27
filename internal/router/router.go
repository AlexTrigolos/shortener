package router

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/thanhpk/randstr"

	"github.com/AlexTrigolos/shortener/internal/handler"
	"github.com/AlexTrigolos/shortener/internal/service"
	api_handler "github.com/AlexTrigolos/shortener/internal/handler/api"
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

	// middleware.Logger убрал втоматическое логирование
	r.Use(middleware.ClientIPFromRemoteAddr, middleware.Recoverer)

	generateURL := service.NewGenerateURL(randstr.String, gr.newURL)
	compact := handler.NewCompact(generateURL)
	apiCompact := api_handler.NewCompact(generateURL)

	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.StatusHandler)
		r.Post("/", compact.CompactHandler(baseURL))
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handler.GetHandler(gr.newURL))
		})
		r.Post("/api/shorten", apiCompact.CompactHandler(baseURL))
	})

	return r, nil
}
