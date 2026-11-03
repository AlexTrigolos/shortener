package main

import (
	"os"

	"github.com/AlexTrigolos/shortener/internal/app"
)

func main() {
	app := app.NewApp()
	err := app.RunApp()
	if err != nil {
		os.Exit(1)
	}
}
