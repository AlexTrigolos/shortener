package config

import (
	"log/slog"
	"net/http"
	"os"
)

type Server struct {
	ListenAndServe func(addr string, handler http.Handler) error
	Flags          Config
	Logger         *slog.Logger
	LoggingURI		 func(h http.Handler) http.Handler
}

func PrepareServer() (*Server, error) {
	flags, err := ParseFlags()
	if err != nil {
		return &Server{Logger: slog.New(slog.NewTextHandler(os.Stdout, nil))}, err
	}
	server := &Server{
		ListenAndServe: http.ListenAndServe,
		Flags:          flags,
		Logger:         slog.New(slog.NewTextHandler(os.Stdout, nil)),
		LoggingURI:			WithLogging,
	}
	return server, nil
}
