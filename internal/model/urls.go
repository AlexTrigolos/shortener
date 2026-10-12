package model

import (
	"errors"
	"sync"

	"github.com/AlexTrigolos/shortener/internal/interfaces"
)

type URL struct {
	mu   sync.Mutex
	data map[string]string
}

func NewURL() (interfaces.GetterSetter, error) {
	return &URL{data: make(map[string]string)}, nil
}

func (s *URL) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, exists := s.data[key]
	if exists {
		return errors.New("такой ключ уже занят")
	}
	s.data[key] = value
	return nil
}

func (s *URL) Get(key string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, exists := s.data[key]
	if !exists {
		return "", errors.New("такой ключ не найден")
	}
	return value, nil
}
