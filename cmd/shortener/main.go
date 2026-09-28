package main

import (
	"net/http"

	"github.com/AlexTrigolos/shortener/internal/router"
)

var (
	listenAndServe = http.ListenAndServe
	generateRouter = router.GenerateRouter
)

func errPanic(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	router, err := generateRouter()
	errPanic(err)

	err = listenAndServe(":8080", router)
	errPanic(err)
}
