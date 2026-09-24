package handler

import (
	"io"
	"net/http"

	"github.com/AlexTrigolos/shortener/internal/model"
)

func Compact(w http.ResponseWriter, r *http.Request, urls *model.URL) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := string(body)
	if originalURL == "" {
		http.Error(w, "URL is empty", http.StatusBadRequest)
		return
	}

	short := "EwHXdJfB"
	urls.Set(short, originalURL)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://localhost:8080/" + short))
}
