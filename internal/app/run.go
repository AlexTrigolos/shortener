package app

import (
	"fmt"
	"net/http"

	"github.com/AlexTrigolos/shortener/internal/handler"
	"github.com/AlexTrigolos/shortener/internal/model"
)

var(
	newURL = model.NewURL
	listenAndServe = http.ListenAndServe
)

func Run() error {
	urls, err := newURL()
	if err != nil {
		fmt.Println("не удалось содать урлы")
		return err
	}
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handler.RootHandler(urls))
	mux.HandleFunc(`/{id}`, handler.GetHandler(urls))

	return listenAndServe(`:8080`, mux)
}
