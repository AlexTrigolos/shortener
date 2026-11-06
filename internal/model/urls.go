package model

import (
	"context"
	"errors"
	"sync"
)

type URL struct {
	mu   sync.Mutex
	data map[string]string
}

func NewURL() (*URL, error) {
	return &URL{data: make(map[string]string)}, nil
}

func (u *URL) Set(_ context.Context, key, value string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	_, exists := u.data[key]
	if exists {
		return errors.New("такой ключ уже занят")
	}
	u.data[key] = value
	return nil
}

func (u *URL) Get(_ context.Context, key string) (string, error) {
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

func (u *URL) PingContext(ctx context.Context) error {
	return errors.New("нет доступа к бд")
}
