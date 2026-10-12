package handler

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/AlexTrigolos/shortener/internal/interfaces"
)

type GenerateAndSetUniqURL func(interfaces.GetterSetter, string) (string, error)

type Compact struct {
	generateAndSetUniqURL GenerateAndSetUniqURL
	logger                *slog.Logger
}

func NewCompact(guu GenerateAndSetUniqURL) *Compact {
	return &Compact{generateAndSetUniqURL: guu, logger: slog.New(slog.NewTextHandler(os.Stdout, nil))}
}

func (c Compact) CompactHandler(urls interfaces.GetterSetter, baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.SetDefault(c.logger)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "не удалось прочитать тело запроса", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		originalURL := string(body)
		if originalURL == "" {
			http.Error(w, "переданный вами URL пустой", http.StatusBadRequest)
			return
		}

		short, err := c.generateAndSetUniqURL(urls, originalURL)
		if err != nil {
			c.logger.Error(fmt.Sprintf("создание короткой ссылки: %s", err))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(baseURL + "/" + short))
	}
}
