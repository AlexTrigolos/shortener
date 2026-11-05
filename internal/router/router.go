package router

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/thanhpk/randstr"

	"github.com/AlexTrigolos/shortener/internal/config"
	"github.com/AlexTrigolos/shortener/internal/handler"
	api_handler "github.com/AlexTrigolos/shortener/internal/handler/api"
	"github.com/AlexTrigolos/shortener/internal/service"
)

type GetterSetter interface {
	Set(key, value string) error
	Get(key string) (string, error)
}

type DB interface {
	PingContext(ctx context.Context) error
}

type GenerateRouter struct {
	db DB
	newURL GetterSetter
}

func NewGenerateRouter(db DB, nu GetterSetter) *GenerateRouter {
	return &GenerateRouter{db: db, newURL: nu}
}

func (gr *GenerateRouter) GenerateRoute(baseURL string, server *config.Server) (chi.Router, error) {
	r := chi.NewRouter()

	// middleware.Logger убрал втоматическое логирование
	r.Use(middleware.ClientIPFromRemoteAddr, server.GzipMiddleware, server.LoggerMiddleware, middleware.Recoverer)

	generateURL := service.NewGenerateURL(randstr.String, gr.newURL)
	compact := handler.NewCompact(generateURL, server.Logger)
	apiCompact := api_handler.NewCompact(generateURL, server.Logger)
	ping := handler.NewPing(gr.db, server.Logger)

	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.StatusHandler)
		r.Post("/", compact.CompactHandler(baseURL))
		r.Get("/ping", ping.PingHandler())
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handler.GetHandler(gr.newURL))
		})
		r.Post("/api/shorten", apiCompact.CompactHandler(baseURL))
	})

	return r, nil
}
