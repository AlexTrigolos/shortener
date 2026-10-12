package main

import (
	"fmt"
	"os"

	"github.com/AlexTrigolos/shortener/internal/config"
	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/AlexTrigolos/shortener/internal/router"
)

func errExitWithLog(server *config.Server, err error) {
	if err != nil {
		server.Logger.Error(fmt.Sprintf("запуск сервера: %s", err))
		os.Exit(1)
	}
}

var prepareServer = config.PrepareServer

func main() {
	server := prepareServer()
	flags := server.Flags

	r := router.NewGenerateRouter(model.NewURL)
	rtr, err := r.GenerateRoute(flags.GetBaseURL())
	errExitWithLog(server, err)

	err = server.ListenAndServe(flags.GetServerAddr(), rtr)
	errExitWithLog(server, err)
}
