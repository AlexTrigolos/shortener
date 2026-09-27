package main

import "github.com/AlexTrigolos/shortener/internal/app"

var run = app.Run

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}
