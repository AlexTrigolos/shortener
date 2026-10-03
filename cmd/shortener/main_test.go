package main

import (
	"errors"
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

var (
	mockAddr = ""
	mockPrs  = 0
)

func mockSuccessListenAndServe(addr string, handler http.Handler) error {
	mockAddr = addr
	return nil
}

func mockFailListenAndServe(addr string, handler http.Handler) error {
	return errors.New("неудачно настроились и запустились хэндлеры")
}

func mockSuccessGenerateRouter(flagDefaultURL string) (chi.Router, error) {
	return chi.NewRouter(), nil
}

func mockFailGenerateRouter(flagDefaultURL string) (chi.Router, error) {
	return chi.NewRouter(), errors.New("не побежал")
}

func mockPrsFlags() {
	flagRunAddr = ":8080"
	flagDefaultURL = "http://localhost:8000/unknown_short_url"
	mockPrs++
}

func TestMain(t *testing.T) {
	prsFlags = mockPrsFlags
	tests := []struct {
		name               string
		mockListenAndServe func(addr string, handler http.Handler) error
		mockGenerateRouter func(string) (chi.Router, error)
		panic              string
	}{
		{
			name:               "удачный запуск",
			mockListenAndServe: mockSuccessListenAndServe,
			mockGenerateRouter: mockSuccessGenerateRouter,
			panic:              "",
		},
		{
			name:               "ошибка при генерации роутов",
			mockListenAndServe: mockSuccessListenAndServe,
			mockGenerateRouter: mockFailGenerateRouter,
			panic:              "не побежал",
		},
		{
			name:               "ошибка запуска сервера",
			mockListenAndServe: mockFailListenAndServe,
			mockGenerateRouter: mockSuccessGenerateRouter,
			panic:              "неудачно настроились и запустились хэндлеры",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listenAndServe = tt.mockListenAndServe
			generateRouter = tt.mockGenerateRouter
			if tt.panic != "" {
				assert.PanicsWithError(t, tt.panic, main)
			} else {
				assert.NotPanics(t, main)
				assert.Equal(t, flagRunAddr, mockAddr)
				assert.Equal(t, 1, mockPrs)
			}
		})
	}
}
