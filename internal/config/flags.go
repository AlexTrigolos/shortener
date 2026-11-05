package config

import (
	"errors"
	"net/url"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/spf13/pflag"
)

type Config struct {
	Addr            string `env:"SERVER_ADDRESS"`
	BaseURL         string `env:"BASE_URL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DataBase        string `env:"DATABASE_DSN"`
}

func (c *Config) prepareURL() error {
	parsed, err := url.Parse(c.BaseURL)
	if err != nil {
		return errors.New("ошибка парсинга флага")
	} else if parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("флаг не является ссылкой")
	}
	c.BaseURL = parsed.Scheme + "://" + parsed.Host
	return nil
}

func ParseFlags() (Config, error) {
	f := pflag.NewFlagSet("app", pflag.ExitOnError)
	config := Config{}

	f.StringVarP(&config.Addr, "addr", "a", ":8080", "address and port to run server")
	f.StringVarP(&config.BaseURL, "base-url", "b", "http://localhost:8080/unknown_short_url", "host (and port) for short urls")
	f.StringVarP(&config.FileStoragePath, "file-storage-path", "f", "./urls_storage.json", "path to file storage for save urls")
	f.StringVarP(&config.DataBase, "data-base", "d", "", "postgresql connection string")
	f.Parse(os.Args[1:])

	err := env.Parse(&config)
	if err != nil {
		return config, err
	}

	err = config.prepareURL()
	if err != nil {
		return config, err
	}
	return config, nil
}
