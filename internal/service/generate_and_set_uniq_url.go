package service

import (
	"errors"
)

type Setter interface {
	Set(key, value string) error
}

type RandStr func(n int, letters ...string) string

type GenerateURL struct {
	randStr RandStr
}

func NewGenerateURL(rs RandStr) *GenerateURL {
	return &GenerateURL{randStr: rs}
}

func (gu GenerateURL) GenerateAndSetUniqURL(urls Setter, originalURL string) (string, error) {
	for range 10 {
		uniqURL := gu.randStr(8)
		err := urls.Set(uniqURL, originalURL)
		if err == nil {
			return uniqURL, nil
		}
	}

	return "", errors.New("исчерпано количество попыток создать короткую ссылку")
}
