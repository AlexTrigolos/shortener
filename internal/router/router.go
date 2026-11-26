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
	Set(ctx context.Context, key, value, userID string) error
	Get(ctx context.Context, key string) (string, error)
	GetByOriginal(ctx context.Context, original, baseURL string) (string, error)
	GetByUserID(_ context.Context, userID, baseURL string) (*[]model.UserURL, error)

	CheckUserIDURLs(ctx context.Context, urls []string, userID string) error
	DeleteURLs(ctx context.Context, urls []model.URLsData) error

	PingContext(ctx context.Context) error
	Begin(count int) (*model.Transaction, error)
}

type AddDelURLs interface {
	AddURLs(urls []string, userID string)
}

type GenerateRouter struct {
	URLs    GetterSetterWithConflictDBReady
	deleter AddDelURLs
}

func NewGenerateRouter(nu GetterSetterWithConflictDBReady, deleter AddDelURLs) *GenerateRouter {
	return &GenerateRouter{URLs: nu, deleter: deleter}
}

func (gr *GenerateRouter) GenerateRoute(baseURL string, server *config.Server) (chi.Router, error) {
	r := chi.NewRouter()

	r.Use(middleware.ClientIPFromRemoteAddr, server.GzipMiddleware, server.LoggerMiddleware, server.SignedCookieMiddleware, middleware.Recoverer)

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
			r.Get("/", handler.GetHandler(gr.URLs, server.Logger))
		})
		r.Route("/api", func(r chi.Router) {
			r.Get("/user/urls", api_handler.GetUserURLsHandler(baseURL, gr.URLs, server.Logger))
			r.Delete("/user/urls", api_handler.DeleteUserURLsHandler(gr.URLs, gr.deleter, server.Logger))
			r.Route("/shorten", func(r chi.Router) {
				r.Post("/", apiCompact.CompactHandler(baseURL))
				r.Post("/batch", apiBatchCompact.BatchCompactHandler(baseURL))
			})
		})
	})

	return r, nil
}
