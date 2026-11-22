package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/AlexTrigolos/shortener/internal/auth"
	"github.com/AlexTrigolos/shortener/internal/model"
	"go.uber.org/zap"
)

type UserGetter interface {
	GetByUserID(ctx context.Context, userID, baseURL string) (*[]model.UserURL, error)
}

func GetUserURLsHandler(baseURL string, urls UserGetter, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := r.Context().Value(auth.CookieName).(string)
		if !ok {
			logger.Error("не установлено значение user_id в контексте")
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 1*time.Second)
		defer cancel()

		userURLs, err := urls.GetByUserID(ctx, userID, baseURL)
		if err != nil {
			logger.Error("получение списка сокращений пользователя "+userID, zap.String("Error", err.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if len(*userURLs) == 0 {
			http.Error(w, http.StatusText(http.StatusNoContent), http.StatusNoContent)
			return
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		enc := json.NewEncoder(w)
		if err := enc.Encode(*userURLs); err != nil {
			logger.Error("сжатие ответа", zap.String("Error", err.Error()))
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}
