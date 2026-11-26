package middleware

import (
	"context"
	"net/http"

	"github.com/AlexTrigolos/shortener/internal/auth"
	"github.com/thanhpk/randstr"
	"go.uber.org/zap"
)

type SignedCookie struct {
	signer *auth.Signer
	logger *zap.Logger
}

func NewSignedCookie(signer *auth.Signer, logger *zap.Logger) *SignedCookie {
	return &SignedCookie{signer: signer, logger: logger}
}

func (sc SignedCookie) SignedCookieMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string

		cookie, err := r.Cookie(string(auth.CookieName))
		if err == nil {
			id, ok := sc.signer.Verify(cookie.Value)
			if ok {
				userID = id
				sc.logger.Info("кука пользователя", zap.String("id", userID))
			} else {
				sc.logger.Info("кука пользователя не прошла проверку подлинности, выдаём новый id")
			}
		} else {
			sc.logger.Info("нет куки, выдаём новый id")
		}

		if userID == "" {
			userID = randstr.Base62(22)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     string(auth.CookieName),
			Value:    sc.signer.Sign(userID),
			Path:     "/",
			MaxAge:   auth.CookieMaxAge,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   true,
		})

		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), auth.CookieName, userID)))
	})
}
