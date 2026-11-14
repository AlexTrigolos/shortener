package service

import (
	"context"
	"errors"

	"github.com/AlexTrigolos/shortener/internal/model"
)

type BatchSetter interface {
	Begin(count int) (*model.Transaction, error)
}

type BatchGenerateURL struct {
	randStr func(n int, letters ...string) string
	storage BatchSetter
}

func NewBatchGenerateURL(rs func(n int, letters ...string) string, storage BatchSetter) *BatchGenerateURL {
	return &BatchGenerateURL{randStr: rs, storage: storage}
}

func (gu BatchGenerateURL) BatchGenerateAndSetUniqURL(ctx context.Context, originalURLs *[]model.BatchRequest, baseURL string) (*[]model.BatchResponse, error) {
	shortLen := 8
	randBatch := 100
	generateCount := shortLen * randBatch
	// для оптимизации создаем длинную строку из которой последовательно получаем случайный набор символов длинной shortLen, генерируем для randBatch числа уникальных строк
	generatedStr := gu.randStr(generateCount)
	startStr := 0
	endStr := shortLen
	result := make([]model.BatchResponse, 0, len(*originalURLs))
	maxRetries := 10
	var retryNum int

	tx, err := gu.storage.Begin(len(*originalURLs))
	if err != nil {
		return &result, err
	}
	defer tx.Rollback()

	for _, originalURLInfo := range *originalURLs {
		for retryNum = 0; retryNum < maxRetries; retryNum++ {
			if endStr == generateCount {
				generatedStr = gu.regenStr(generateCount)
				startStr = 0
				endStr = shortLen
			}
			uniqURL := generatedStr[startStr:endStr]
			startStr += shortLen
			endStr += shortLen
			err := tx.SetTx(ctx, uniqURL, originalURLInfo.OriginalURL)
			if err == nil {
				result = append(result, model.BatchResponse{CorrelationID: originalURLInfo.CorrelationID, ShortURL: baseURL + "/" + uniqURL})
				break
			}
		}
		if retryNum == maxRetries {
			return &result, errors.New("исчерпано количество попыток создать короткую ссылку")
		}
	}
	err = tx.Commit()
	if err != nil {
		return &result, err
	}
	return &result, nil
}

func (gu BatchGenerateURL) regenStr(generateCount int) string {
	return gu.randStr(generateCount)
}
