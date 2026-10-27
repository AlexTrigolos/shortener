package handler

import (
	"io"
	"log/slog"
	"net/http"
	"os"
)

type URLGenerator interface {
	GenerateAndSetUniqURL(originalURL string) (string, error)
}

type Compact struct {
	generator URLGenerator
	logger    *slog.Logger
}

func NewCompact(ug URLGenerator) *Compact {
	return &Compact{generator: ug, logger: slog.New(slog.NewTextHandler(os.Stdout, nil))}
}

func (c Compact) CompactHandler(baseURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		short, err := c.generator.GenerateAndSetUniqURL(originalURL)
		if err != nil {
			c.logger.Error("создание короткой ссылки", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(baseURL + "/" + short))
	}
}
