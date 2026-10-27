package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/AlexTrigolos/shortener/internal/model"
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
		var req model.Request
		dec := json.NewDecoder(r.Body)
		if err := dec.Decode(&req); err != nil {
			_, err := io.ReadAll(r.Body)
			defer r.Body.Close()

			if err != nil {
				http.Error(w, "не удалось прочитать тело запроса", http.StatusBadRequest)
				return
			}

			http.Error(w, "не можем декодировать JSON тело запроса", http.StatusBadRequest)
			return
		}

		originalURL := string(req.URL)
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		resp := model.Response{Result: baseURL + "/" + short}
		enc := json.NewEncoder(w)
		if err := enc.Encode(resp); err != nil {
			c.logger.Error("создание короткой ссылки", "error", err)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}
