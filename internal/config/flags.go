package config

import (
	"log/slog"
	"os"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/spf13/pflag"
)

type Config struct {
	Addr    string `env:"SERVER_ADDRESS"`
	BaseURL string `env:"BASE_URL"`
}

func (c *Config) prepareURL() {
	if c.BaseURL != "" {
		c.BaseURL = strings.Join(strings.Split(c.BaseURL, "/")[:3], "/")
	}
}

func (c Config) GetServerAddr() string {
	return c.Addr
}

func (c Config) GetBaseURL() string {
	return c.BaseURL
}

func ParseFlags() Config {
	f := pflag.NewFlagSet("app", pflag.ExitOnError)
	config := Config{}

	f.StringVarP(&config.Addr, "addr", "a", ":8080", "address and port to run server")
	f.StringVarP(&config.BaseURL, "base-url", "b", "http://localhost:8080/unknown_short_url", "host (and port) for short urls")
	f.Parse(os.Args[1:])

	err := env.Parse(&config)
	if err != nil {
		logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
		logger.Error("не удалось распарсить переменные окружения", "error", err)
	}

	config.prepareURL()

	return config
}
