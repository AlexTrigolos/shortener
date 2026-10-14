package config

import (
	"os"
	"strings"

	"github.com/spf13/pflag"
)

type Config struct {
	addr    string
	baseURL string
}

func (c *Config) prepareURL() {
	c.baseURL = strings.Join(strings.Split(c.baseURL, "/")[:3], "/")
}

func (c Config) GetServerAddr() string {
	return c.addr
}

func (c Config) GetBaseURL() string {
	return c.baseURL
}

func ParseFlags() Config {
	f := pflag.NewFlagSet("app", pflag.ExitOnError)
	config := Config{}
	f.StringVarP(&config.addr, "addr", "a", ":8080", "address and port to run server")
	f.StringVarP(&config.baseURL, "base-url", "b", "http://localhost:8080/unknown_short_url", "host (and port) for short urls")
	f.Parse(os.Args[1:])
	config.prepareURL()
	return config
}
