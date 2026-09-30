package handler

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/go-chi/chi/v5"
)

var (
	httpRedirect = http.Redirect
	logger       = slog.New(slog.NewTextHandler(os.Stdout, nil))
)

func init() {
	slog.SetDefault(logger)
}

func GetHandler(urls *model.URL, flagDefaultURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		url, err := urls.Get(id)
		if err != nil {
			slog.Info("Не нашли запись для: " + id)
			url = flagDefaultURL
		}

		httpRedirect(w, r, url, http.StatusTemporaryRedirect)
	}
}
