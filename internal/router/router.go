package router

import (
	"context"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/thanhpk/randstr"

	"github.com/AlexTrigolos/shortener/internal/config"
	"github.com/AlexTrigolos/shortener/internal/handler"
	api_handler "github.com/AlexTrigolos/shortener/internal/handler/api"
	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/AlexTrigolos/shortener/internal/service"
)

type GetterSetterWithConflictDBReady interface {
	Set(ctx context.Context, key, value string) error
	Get(ctx context.Context, key string) (string, error)
	GetByOriginal(ctx context.Context, original string) (string, error)

	PingContext(ctx context.Context) error
	Begin(count int) (*model.Transaction, error)
	Close() error
}

type GenerateRouter struct {
	URLs GetterSetterWithConflictDBReady
}

func NewGenerateRouter(nu GetterSetterWithConflictDBReady) *GenerateRouter {
	return &GenerateRouter{URLs: nu}
}

func (gr *GenerateRouter) GenerateRoute(baseURL string, server *config.Server) (chi.Router, error) {
	r := chi.NewRouter()

	// middleware.Logger убрал втоматическое логирование
	r.Use(middleware.ClientIPFromRemoteAddr, server.GzipMiddleware, server.LoggerMiddleware, middleware.Recoverer)

	generateURL := service.NewGenerateURL(randstr.String, gr.URLs)
	compact := handler.NewCompact(generateURL, server.Logger)
	apiCompact := api_handler.NewCompact(generateURL, server.Logger)

	batchGenerateURL := service.NewBatchGenerateURL(randstr.String, gr.URLs)
	apiBatchCompact := api_handler.NewBatchCompact(batchGenerateURL, server.Logger)

	ping := handler.NewPing(gr.URLs, server.Logger)

	r.Route("/", func(r chi.Router) {
		r.Get("/", handler.StatusHandler)
		r.Post("/", compact.CompactHandler(baseURL))
		r.Get("/ping", ping.PingHandler())
		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", handler.GetHandler(gr.URLs))
		})
		r.Route("/api/shorten", func(r chi.Router) {
			r.Post("/", apiCompact.CompactHandler(baseURL))
			r.Post("/batch", apiBatchCompact.BatchCompactHandler(baseURL))
		})
	})

	return r, nil
}
