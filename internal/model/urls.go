package model

import (
	"errors"
	"sync"
)

type URL struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewURL() (*URL, error) {
	return &URL{data: make(map[string]string)}, nil
}

func (s *URL) IsExists(key string) bool {
	_, exists := s.data[key]
	return exists
}

func (s *URL) Set(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	exists := s.IsExists(key)
	if exists {
		return errors.New("такой ключ уже занят")
	}
	s.data[key] = value
	return nil
}

func (s *URL) Get(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	exists := s.IsExists(key)
	if !exists {
		return "", errors.New("такой ключ не найден")
	}
	return s.data[key], nil
}
