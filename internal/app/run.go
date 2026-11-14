package app

import (
	"database/sql"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/AlexTrigolos/shortener/internal/config"
	migrateDB "github.com/AlexTrigolos/shortener/internal/db"
	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/AlexTrigolos/shortener/internal/router"
	"github.com/golang-migrate/migrate/v4/database"
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

	var URLs router.GetterSetterWithConflictDBReady
	var URLErr error
	if flags.DataBase != "" {
		db, err := sql.Open("pgx", flags.DataBase)
		if err != nil {
			return logAndReturnError(logger, err)
		}
		defer db.Close()

		var dbDrv *database.Driver
		var m *migrate.Migrate

		URLs, dbDrv, m, URLErr = model.NewDBURL(db, &migrateDB.MigrationService{})

		defer func() {
			if dbDrv != nil {
				(*dbDrv).Close()
			}
			if m != nil {
				m.Close()
			}
		}()
	} else {
		if flags.FileStoragePath != "" {
			URLs, URLErr = model.NewURLWithFile(flags.FileStoragePath)
		} else {
			URLs, URLErr = model.NewURL()
		}
	}
	if URLErr != nil {
		return logAndReturnError(logger, URLErr)
	}
	defer func() {
		err := URLs.Close()
		if err != nil {
			logAndReturnError(logger, err)
		}
	}()
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
