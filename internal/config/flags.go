package config

import (
	"errors"
	"net/url"
	"os"

	"github.com/spf13/pflag"
)

type Config struct {
	addr    string
	baseURL string
}

func (c *Config) prepareURL() error {
	parsed, err := url.Parse(c.baseURL)
	if err != nil {
		return errors.New("ошибка парсинга флага")
	} else if parsed.Scheme == "" || parsed.Host == "" {
		return errors.New("флаг не является ссылкой")
	}
	c.baseURL = parsed.Scheme + "://" + parsed.Host
	return nil
}

func (c Config) GetServerAddr() string {
	return c.addr
}

func (c Config) GetBaseURL() string {
	return c.baseURL
}

func ParseFlags() (Config, error) {
	f := pflag.NewFlagSet("app", pflag.ExitOnError)
	config := Config{}
	f.StringVarP(&config.addr, "addr", "a", ":8080", "address and port to run server")
	f.StringVarP(&config.baseURL, "base-url", "b", "http://localhost:8080/unknown_short_url", "host (and port) for short urls")
	f.Parse(os.Args[1:])
	err := config.prepareURL()
	if err != nil {
		return config, err
	}
	return config, nil
}
