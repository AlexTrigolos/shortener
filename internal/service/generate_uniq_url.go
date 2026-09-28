package service

import (
	"github.com/AlexTrigolos/shortener/internal/model"
	"github.com/thanhpk/randstr"
)

var randStr = randstr.String

func GenerateUniqURL(urls *model.URL) string {
	uniqURL := randStr(8)

	for {
		_, err := urls.Get(uniqURL)
		if err != nil {
			break
		}
		uniqURL = randStr(8)
	}

	return uniqURL
}
