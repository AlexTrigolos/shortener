package model

import (
	"context"
	"errors"
	"sync"
)

type URL struct {
	mu                sync.Mutex
	data              map[string]string
	transactionBuffer map[string]string
	values            map[string]string
}

func NewURL() (*URL, error) {
	return &URL{data: make(map[string]string), values: make(map[string]string)}, nil
}

func (u *URL) Set(_ context.Context, key, value string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	_, exists := u.values[value]
	if exists {
		return ErrConflict
	}
	_, exists = u.data[key]
	if exists {
		return errors.New("такой ключ уже занят")
	}
	u.values[value] = key
	u.data[key] = value
	return nil
}

func (u *URL) Get(_ context.Context, key string) (string, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	value, exists := u.data[key]
	if !exists {
		return "", errors.New("не найден ключ " + key)
	}
	return value, nil
}

func (u *URL) GetByOriginal(_ context.Context, original string) (string, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	short, exists := u.values[original]
	if !exists {
		return "", errors.New("не нашли значение " + original)
	}
	return short, nil
}

func (u *URL) delete(key string) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.values, u.data[key])
	delete(u.data, key)
	return nil
}

func (u *URL) PingContext(ctx context.Context) error {
	return errors.New("нет доступа к бд")
}

// имитируем транзакцию
func (u *URL) Begin(count int) (*Transaction, error) {
	u.mu.Lock()
	u.transactionBuffer = make(map[string]string, count)
	transaction, err := NewTransaction(u, nil, u)
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

func (u *URL) Rollback() error {
	defer u.mu.Unlock()
	u.transactionBuffer = nil
	return nil
}

func (u *URL) Commit() error {
	defer u.mu.Unlock()
	defer func() { u.transactionBuffer = nil }()

	savedKeys := []string{}
	for k, v := range u.transactionBuffer {
		_, exists := u.data[k]
		if exists {
			for _, delKey := range savedKeys {
				delete(u.values, u.data[delKey])
				delete(u.data, delKey)
			}
			return errors.New(k + " ключ уже занят")
		}
		_, exists = u.values[v]
		if exists {
			for _, delKey := range savedKeys {
				delete(u.values, u.data[delKey])
				delete(u.data, delKey)
			}
			return errors.New(k + " значение уже занято")
		}
		u.values[v] = k
		u.data[k] = v
		savedKeys = append(savedKeys, k)
	}
	return nil
}

func (u *URL) SetTx(_ context.Context, key, value string) error {
	_, exists := u.data[key]
	if exists {
		return errors.New(key + " ключ уже занят")
	}
	_, exists = u.transactionBuffer[key]
	if exists {
		return errors.New(key + " ключ уже занят в транзакции")
	}
	u.transactionBuffer[key] = value
	return nil
}

func (u *URL) Close() error {
	return nil
}
