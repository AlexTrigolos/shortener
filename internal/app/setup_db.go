package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// как то не очень хочу на эту жепу делать тесты, ожидается сильная сложнота, пощадитъэ
func setupDB(ctx context.Context, dataBase string) (*sql.DB, error) {
	// подключаемся к бд
	targetDB, err := sql.Open("pgx", dataBase)
	if err != nil {
		return nil, err
	}

	// если пинга нету, необходимо создать бд
	err = targetDB.Ping()
	if err != nil {
		var connectError *pgconn.ConnectError
		// по идее если нет бд то тут мы игнорим ошибку
		if !errors.As(err, &connectError) {
			return nil, err
		}

		adminConnStr := createAdminConnString(dataBase)
		adminDB, err := sql.Open("pgx", adminConnStr)
		// подключаемся к БД postgres по идее она должна быть по умолчанию
		if err != nil {
			return nil, err
		}
		defer adminDB.Close()

		// из неё создаем нашу бд
		_, err = adminDB.ExecContext(ctx, "CREATE DATABASE "+getDBName(dataBase))
		if err != nil {
			return nil, err
		}

		targetDB, err = sql.Open("pgx", dataBase)
		if err != nil {
			return nil, err
		}

		// перешли в неё и перепроверили, что все ок
		err = targetDB.Ping()
		if err != nil {
			return nil, err
		}
	}
	// на всякий говорим, чтоб закрылась
	defer targetDB.Close()

	// запускаем миграции
	if err := runMigrations(targetDB); err != nil {
		return nil, err
	}

	targetDB.Close()

	// мы хотим `defer db.Close()` делать в той же функции, что и запус сервера, поэтому тут принулительно переоткрываем, чтоб вернуть
	db, err := sql.Open("pgx", dataBase)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// Вспомогательная функция для удаления dbname из строки
func createAdminConnString(dsn string) string {
	// Простой способ: заменить dbname на postgres
	return strings.Replace(dsn, "dbname="+getDBName(dsn), "dbname=postgres", 1)
}

// получаем название бд
func getDBName(dsn string) string {
	parts := strings.Split(dsn, " ")
	for _, p := range parts {
		if strings.HasPrefix(p, "dbname=") {
			return strings.TrimPrefix(p, "dbname=")
		}
	}
	return ""
}

func runMigrations(db *sql.DB) error {
	// драйвер для выполнения миграций
	driver, err := pgx.WithInstance(db, &pgx.Config{})
	if err != nil {
		return err
	}

	// получаем путь к root проекта
	root, err := getProjectRoot()
	if err != nil {
		return err
	}

	// для нормального считывания обрабатываем если запуск на винде + показываем, что миграции у нас в файле + путь до них
	migratePath := windowsTransform(fmt.Sprintf("file://%s", filepath.Join(root, "internal", "migrations")))
	m, err := migrate.NewWithDatabaseInstance(migratePath, "postgres", driver)
	if err != nil {
		return err
	}

	// запускаем миграции и смотрим, что ошибка не отсутствие изменений
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func getProjectRoot() (string, error) {
	// Начинаем поиск с директории, где лежит текущий исполняемый файл
	currentPath, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Идем вверх по папкам, пока не найдем go.mod
	for {
		path := filepath.Join(currentPath, "go.mod")
		if _, err := os.Stat(path); err == nil {
			return currentPath, nil // Нашли!
		}

		// Если дошли до корня диска (/ или C:$, останавливаемся
		parent := filepath.Dir(currentPath)
		if parent == currentPath {
			break
		}
		currentPath = parent
	}
	return "", os.ErrNotExist
}

func windowsTransform(winPath string) string {
	uriPath := filepath.ToSlash(winPath)
	if !strings.HasPrefix(uriPath, "/") && len(uriPath) > 1 && uriPath[1] == ':' {
		uriPath = "/" + uriPath
	}
	return uriPath
}
