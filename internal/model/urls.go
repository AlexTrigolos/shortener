package model

import (
	"errors"
	"sync"
)

type URL struct {
	mu   sync.Mutex
	data map[string]string
}

func NewURL() (*URL, error) {
	url := &URL{data: make(map[string]string)}
	return url, nil
}

func (u *URL) Set(key, value string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	_, exists := u.data[key]
	if exists {
		return errors.New("такой ключ уже занят")
	}
	u.data[key] = value
	return nil
}

func (u *URL) Get(key string) (string, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	value, exists := u.data[key]
	if !exists {
		return "", errors.New("такой ключ не найден")
	}
	return value, nil
}

func (u *URL) delete(key string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.data, key)
	return nil
}
