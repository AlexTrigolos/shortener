package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/AlexTrigolos/shortener/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const cookieSecret = "secret_key"

func TestNewSignedCookie(t *testing.T) {
	signer := auth.NewSigner("secret_key")
	logger := zap.NewNop()

	t.Run("создать новый SignedCookie", func(t *testing.T) {
		newSignedCookie := NewSignedCookie(signer, logger)

		expectedSignedCookie := &SignedCookie{signer: signer, logger: logger}
		assert.Equal(t, expectedSignedCookie, newSignedCookie)
	})
}

func TestSignedCookie_SignedCookieMiddleware(t *testing.T) {
	signer := auth.NewSigner(cookieSecret)
	signedCookie := NewSignedCookie(signer, zap.NewNop())

	var gotUserID string
	handler := signedCookie.SignedCookieMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = r.Context().Value(auth.CookieName).(string)
		w.WriteHeader(http.StatusOK)
	}))
	srv := httptest.NewServer(handler)
	defer srv.Close()

	badCookie := "notAValidSignedValue"
	validCookie := signer.Sign("valid_user_cookie")

	tests := []struct {
		name          string
		requestCookie string
	}{
		{
			name:          "нет куки",
			requestCookie: "",
		},
		{
			name:          "есть куки без верной проверки",
			requestCookie: string(auth.CookieName) + "=" + badCookie,
		},
		{
			name:          "есть полностью валидный куки",
			requestCookie: string(auth.CookieName) + "=" + validCookie,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", srv.URL, nil)
			r.RequestURI = ""
			r.Header.Set("Cookie", tt.requestCookie)

			resp, err := http.DefaultClient.Do(r)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, http.StatusOK, resp.StatusCode)

			cookie := resp.Header.Get("Set-Cookie")

			switch tt.name {
			case "есть полностью валидный куки":
				require.Equal(t, validCookie[:strings.LastIndex(validCookie, ".")], gotUserID)
				assert.Contains(t, cookie, "shorten_service="+validCookie+"; Path=/; Max-Age="+strconv.Itoa(auth.CookieMaxAge)+"; HttpOnly; Secure; SameSite=Lax")
			case "есть куки без верной проверки":
				require.NotContains(t, cookie, badCookie)
				fallthrough
			case "нет куки":
				assert.Contains(t, cookie, "shorten_service="+signer.Sign(gotUserID)+"; Path=/; Max-Age="+strconv.Itoa(auth.CookieMaxAge)+"; HttpOnly; Secure; SameSite=Lax")
			default:
				require.Fail(t, "unexpected test case")
			}
		})
	}
}
