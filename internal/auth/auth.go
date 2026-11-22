package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)
type AuthString string
const CookieName AuthString = "shorten_service"
const CookieMaxAge = 60 * 60 * 24 * 30

type Signer struct {
	key []byte
}

func NewSigner(secret string) *Signer {
	return &Signer{key: []byte(secret)}
}

func (s *Signer) Sign(value string) string {
	m := hmac.New(sha256.New, s.key)
	m.Write([]byte(value))
	return value + "." + hex.EncodeToString(m.Sum(nil))
}

func (s *Signer) Verify(signed string) (string, bool) {
	i := strings.LastIndex(signed, ".")
	if i <= 0 || i == len(signed)-1 {
		return "", false
	}
	value, signature := signed[:i], signed[i+1:]

	signatureData, err := hex.DecodeString(signature)
	if err != nil {
		return "", false
	}

	h := hmac.New(sha256.New, s.key)
	h.Write([]byte(value))
	sign := h.Sum(nil)

	if !hmac.Equal(sign, signatureData) {
		return "", false
	}
	return string(value), true
}

func (s *Signer) Value(signed string) string {
	i := strings.LastIndex(signed, ".")
	if i <= 0 || i == len(signed)-1 {
		return ""
	}

	return signed[:i]
}
