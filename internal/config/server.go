package config

import (
	"net/http"

	"github.com/AlexTrigolos/shortener/internal/auth"
	"github.com/AlexTrigolos/shortener/internal/config/middleware"
	"go.uber.org/zap"
)

type Server struct {
	ListenAndServe         func(addr string, handler http.Handler) error
	Flags                  Config
	Logger                 *zap.Logger
	LoggerMiddleware       func(h http.Handler) http.Handler
	GzipMiddleware         func(h http.Handler) http.Handler
	SignedCookieMiddleware func(h http.Handler) http.Handler
}

func PrepareServer() (*Server, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return &Server{}, err
	}

	flags, err := ParseFlags()
	if err != nil {
		return &Server{Logger: logger}, err
	}

	signedCookie := middleware.NewSignedCookie(auth.NewSigner(flags.SecretKey), logger)

	server := &Server{
		ListenAndServe:         http.ListenAndServe,
		Flags:                  flags,
		Logger:                 logger,
		LoggerMiddleware:       middleware.LoggerMiddleware(logger),
		GzipMiddleware:         middleware.GzipMiddleware,
		SignedCookieMiddleware: signedCookie.SignedCookieMiddleware,
	}
	return server, nil
}
