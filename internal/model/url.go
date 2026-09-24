package model

import "sync"

type URL struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewURL() *URL {
	return &URL{data: make(map[string]string)}
}

func (s *URL) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *URL) Get(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}
