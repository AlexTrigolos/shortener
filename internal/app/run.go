package app

import (
	"log/slog"
	"os"

	"github.com/AlexTrigolos/shortener/internal/config"
	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/AlexTrigolos/shortener/internal/router"
	"go.uber.org/zap"
)

func errExitWithLog(logger *zap.Logger, err error) {
	if err != nil {
		logger.Error("запуск сервера", zap.String("Error", err.Error()))
		os.Exit(1)
	}
}

func invalidLogger(logger *zap.Logger, err error) {
	if logger == nil {
		lg := slog.New(slog.NewTextHandler(os.Stdout, nil))
		lg.Error("не удалось создать zap логер", "error", err)
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
	logger := server.Logger
	invalidLogger(server.Logger, err)
	errExitWithLog(logger, err)
	flags := server.Flags

	newURLWithFile, err := model.NewURLWithFile(server.Flags.FileStoragePath)
	errExitWithLog(logger, err)
	r := router.NewGenerateRouter(newURLWithFile)

	rtr, err := r.GenerateRoute(flags.GetBaseURL(), server)
	errExitWithLog(logger, err)

	err = server.ListenAndServe(flags.GetServerAddr(), rtr)
	errExitWithLog(logger, err)
}
