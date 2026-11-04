package app

import (
	"log/slog"
	"os"

	"github.com/AlexTrigolos/shortener/internal/config"
	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/AlexTrigolos/shortener/internal/router"
	"go.uber.org/zap"
)

func logAndReturnError(logger *zap.Logger, err error) error {
	logger.Error("запуск сервера", zap.String("Error", err.Error()))
	return err
}

func invalidLogger(logger *zap.Logger, err error) bool {
	if logger != nil {
		return false
	}

	lg := slog.New(slog.NewTextHandler(os.Stdout, nil))
	lg.Error("не удалось создать zap логер", "error", err)
	return true
}

type App struct {
	prepareServer func() (*config.Server, error)
}

func NewApp() *App {
	return &App{prepareServer: config.PrepareServer}
}

func (a App) RunApp() error {
	server, err := a.prepareServer()

	logger := server.Logger
	if invalidLogger(logger, err) {
		return err
	}
	defer logger.Sync()

	if err != nil {
		return logAndReturnError(logger, err)
	}

	flags := server.Flags

	newURLWithFile, err := model.NewURLWithFile(server.Flags.FileStoragePath)
	if err != nil {
		return logAndReturnError(logger, err)
	}
	r := router.NewGenerateRouter(newURLWithFile)

	rtr, err := r.GenerateRoute(flags.GetBaseURL(), server)
	if err != nil {
		return logAndReturnError(logger, err)
	}

	err = server.ListenAndServe(flags.GetServerAddr(), rtr)
	if err != nil {
		return logAndReturnError(logger, err)
	}
	return nil
}
