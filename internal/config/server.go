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
}

func PrepareServer() *Server {
	server := &Server{
		ListenAndServe: http.ListenAndServe,
		Flags:          ParseFlags(),
		Logger:         slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
	slog.SetDefault(server.Logger)
	return server
}
