package service

import (
	"context"
	"errors"

	"github.com/AlexTrigolos/shortener/internal/model"
)

type SetterWithConflict interface {
	Set(ctx context.Context, key, value string) error
	GetByOriginal(ctx context.Context, original string) (string, error)
}

type GenerateURL struct {
	randStr func(n int, letters ...string) string
	storage SetterWithConflict
}

func NewGenerateURL(rs func(n int, letters ...string) string, storage SetterWithConflict) *GenerateURL {
	return &GenerateURL{randStr: rs, storage: storage}
}

func (gu GenerateURL) GenerateAndSetUniqURL(ctx context.Context, originalURL string) (string, error) {
	for range 10 {
		uniqURL := gu.randStr(8)
		err := gu.storage.Set(ctx, uniqURL, originalURL)
		if err == nil {
			return uniqURL, nil
		} else if errors.Is(err, model.ErrConfilct) {
			uniqURL, getErr := gu.storage.GetByOriginal(ctx, originalURL)
			if getErr != nil {
				return "", getErr
			}
			return uniqURL, err
		}
	}

	return "", errors.New("исчерпано количество попыток создать короткую ссылку")
}
