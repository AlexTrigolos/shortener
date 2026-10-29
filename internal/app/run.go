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
	prepareServer func() (*config.Server, error)
}

func NewApp() *App {
	return &App{prepareServer: config.PrepareServer}
}

func (a App) RunApp() {
	server, err := a.prepareServer()
	errExitWithLog(server, err)
	flags := server.Flags

	newURL, err := model.NewURL(server.Flags.FileStoragePath)
	errExitWithLog(server, err)
	r := router.NewGenerateRouter(newURL)

	rtr, err := r.GenerateRoute(flags.GetBaseURL())
	errExitWithLog(server, err)

	err = server.ListenAndServe(flags.GetServerAddr(), server.GzipMiddleware(server.LoggerMiddleware(rtr)))
	errExitWithLog(server, err)
}
