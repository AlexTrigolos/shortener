package service

import (
	"errors"

	"github.com/AlexTrigolos/shortener/internal/interfaces"
)

type RandStr func(n int, letters ...string) string

type GenerateURL struct {
	randStr RandStr
}

func NewGenerateURL(rs RandStr) *GenerateURL {
	return &GenerateURL{randStr: rs}
}

func (gu GenerateURL) GenerateAndSetUniqURL(urls interfaces.GetterSetter, originalURL string) (string, error) {
	for range 10 {
		uniqURL := gu.randStr(8)
		err := urls.Set(uniqURL, originalURL)
		if err == nil {
			return uniqURL, nil
		}
	}

	return "", errors.New("исчерпано количество попыток создать короткую ссылку")
}
