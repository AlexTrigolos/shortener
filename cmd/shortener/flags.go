package main

import (
	"flag"
)

var (
	flagRunAddr    string
	flagDefaultURL string
)

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&flagDefaultURL, "b", "http://localhost:8080/unknown_short_url", "default url for unknown short url")
	flag.Parse()
}
