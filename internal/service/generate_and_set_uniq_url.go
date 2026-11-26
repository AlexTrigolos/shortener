package service

import (
	"context"
	"errors"

	"github.com/AlexTrigolos/shortener/internal/model"
)

type SetterWithConflict interface {
	Set(ctx context.Context, key, value, userID string) error
	GetByOriginal(ctx context.Context, original, baseURL string) (string, error)
}

type GenerateURL struct {
	randStr func(n int, letters ...string) string
	storage SetterWithConflict
}

func NewGenerateURL(rs func(n int, letters ...string) string, storage SetterWithConflict) *GenerateURL {
	return &GenerateURL{randStr: rs, storage: storage}
}

func (gu GenerateURL) GenerateAndSetUniqURL(ctx context.Context, originalURL, userID, baseURL string) (string, error) {
	for range 10 {
		uniqURL := gu.randStr(8)
		err := gu.storage.Set(ctx, uniqURL, originalURL, userID)
		if err == nil {
			return baseURL + "/" + uniqURL, nil
		} else if errors.Is(err, model.ErrConflict) {
			uniqURL, getErr := gu.storage.GetByOriginal(ctx, originalURL, baseURL)
			if getErr != nil {
				return "", getErr
			}
			return uniqURL, err
		}
	}

	return "", errors.New("исчерпано количество попыток создать короткую ссылку")
}
