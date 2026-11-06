package service

import (
	"context"
	"errors"
)

type Setter interface {
	Set(ctx context.Context, key, value string) error
}

type GenerateURL struct {
	randStr func(n int, letters ...string) string
	storage Setter
}

func NewGenerateURL(rs func(n int, letters ...string) string, storage Setter) *GenerateURL {
	return &GenerateURL{randStr: rs, storage: storage}
}

func (gu GenerateURL) GenerateAndSetUniqURL(ctx context.Context, originalURL string) (string, error) {
	for range 10 {
		uniqURL := gu.randStr(8)
		err := gu.storage.Set(ctx, uniqURL, originalURL)
		if err == nil {
			return uniqURL, nil
		}
	}

	return "", errors.New("исчерпано количество попыток создать короткую ссылку")
}
