package app

import (
	"os"

	"github.com/AlexTrigolos/shortener/internal/config"
	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/AlexTrigolos/shortener/internal/router"
)

func errExitWithLog(server *config.Server, err error) {
	if err != nil {
		server.Logger.Error("запуск сервера", "error", err)
		os.Exit(1)
	}
}

type App struct {
	prepareServer func() *config.Server
}

func NewApp() *App {
	return &App{prepareServer: config.PrepareServer}
}

func (a App) RunApp() {
	server := a.prepareServer()
	flags := server.Flags

	newURL, err := model.NewURL()
	errExitWithLog(server, err)
	r := router.NewGenerateRouter(newURL)

	rtr, err := r.GenerateRoute(flags.GetBaseURL())
	errExitWithLog(server, err)

	err = server.ListenAndServe(flags.GetServerAddr(), rtr)
	errExitWithLog(server, err)
}
