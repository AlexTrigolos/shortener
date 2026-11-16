package database

import (
	"database/sql"
	"embed"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed migrations/*
var migrationsFS embed.FS

type MigrationService struct{}

func (*MigrationService) RunMigrations(db *sql.DB) (*database.Driver, *migrate.Migrate, error) {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, nil, err
	}

	dbDrv, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return &dbDrv, nil, err
	}

	m, err := migrate.NewWithInstance("iofs", src, "postgres", dbDrv)
	if err != nil {
		return &dbDrv, m, err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return &dbDrv, m, err
	}
	return &dbDrv, m, nil
}
