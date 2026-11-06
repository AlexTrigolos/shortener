package app

import (
	"context"
	"log/slog"
	"os"
	"time"

	_ "github.com/golang-migrate/migrate/v4/source/file"

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

	var URLs router.GetterSetterDBReady
	if flags.DataBase != "" {
		// хз норм или нет, мб на старт бд и миграции не особо и нужен контекст
		ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
		defer cancel()

		db, e := setupDB(ctx, flags.DataBase)
		if e != nil {
			return logAndReturnError(logger, err)
		}
		defer db.Close()

		URLs, err = model.NewDBURL(db)
	} else {
		if flags.FileStoragePath != "" {
			URLs, err = model.NewURLWithFile(flags.FileStoragePath)
		} else {
			URLs, err = model.NewURL()
		}
	}
	if err != nil {
		return logAndReturnError(logger, err)
	}
	r := router.NewGenerateRouter(URLs)

	rtr, err := r.GenerateRoute(flags.BaseURL, server)
	if err != nil {
		return logAndReturnError(logger, err)
	}

	err = server.ListenAndServe(flags.Addr, rtr)
	if err != nil {
		return logAndReturnError(logger, err)
	}
	return nil
}
