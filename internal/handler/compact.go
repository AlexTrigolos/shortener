package handler

import (
	"io"
	"net/http"

	"github.com/AlexTrigolos/shortener/internal/model"
)

func CompactHandler(w http.ResponseWriter, r *http.Request, urls *model.URL) {
	if r.Method != http.MethodPost {
		http.Error(w, "ожидался Post запрос", http.StatusMethodNotAllowed)
		return
	}

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

	short := "EwHXdJfB" // TODO тут надо бы будет сделать рандомайзер, скорее всего в service GenerateURL
	err = urls.Set(short, originalURL)
	if err != nil {
		http.Error(w, "ошибка добавления URL", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("http://localhost:8080/" + short))
}
