package main

import (
	"net/http"

	"github.com/AlexTrigolos/shortener/internal/router"
)

var (
	listenAndServe = http.ListenAndServe
	generateRouter = router.GenerateRouter
	prsFlags       = parseFlags
)

func errPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	prsFlags()
	router, err := generateRouter(flagDefaultURL)
	errPanic(err)

	err = listenAndServe(flagRunAddr, router)
	errPanic(err)
}
