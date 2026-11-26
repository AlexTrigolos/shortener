package api

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/AlexTrigolos/shortener/internal/auth"
	"go.uber.org/zap"
)

type CheckUserIDURLs interface {
	CheckUserIDURLs(ctx context.Context, urls []string, userID string) error
}

type AddDeleter interface {
	AddURLs(urls []string, userID string)
}

func DeleteUserURLsHandler(urls CheckUserIDURLs, deleter AddDeleter, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req []string
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1024))
		defer r.Body.Close()
		if err := dec.Decode(&req); err != nil {
			var maxErr *http.MaxBytesError
			if errors.As(err, &maxErr) {
				http.Error(w, "тело запроса слишком большое", http.StatusRequestEntityTooLarge)
				return
			}

			_, err := io.ReadAll(r.Body)

			if err != nil {
				http.Error(w, "не удалось прочитать тело запроса", http.StatusBadRequest)
				return
			}

			http.Error(w, "не можем декодировать JSON тело запроса", http.StatusBadRequest)
			return
		}

		userID, ok := r.Context().Value(auth.CookieName).(string)
		if !ok {
			logger.Error("не установлено значение user_id в контексте")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		err := urls.CheckUserIDURLs(ctx, req, userID)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		deleter.AddURLs(req, userID)

		http.Error(w, http.StatusText(http.StatusAccepted), http.StatusAccepted)
	}
}
