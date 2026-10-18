package service

import (
	"errors"
)

type Setter interface {
	Set(key, value string) error
}

type GenerateURL struct {
	randStr func(n int, letters ...string) string
}

func NewGenerateURL(rs func(n int, letters ...string) string) *GenerateURL {
	return &GenerateURL{randStr: rs}
}

func (gu GenerateURL) GenerateAndSetUniqURL(urls interface{}, originalURL string) (string, error) {
	setter, ok := urls.(Setter)
	if !ok {
		return "", errors.New("не был реализован интерфейс для генерации короткой ссылки")
	}
	for range 10 {
		uniqURL := gu.randStr(8)
		err := setter.Set(uniqURL, originalURL)
		if err == nil {
			return uniqURL, nil
		}
	}

	return "", errors.New("исчерпано количество попыток создать короткую ссылку")
}
